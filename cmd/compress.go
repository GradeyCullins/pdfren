package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	compressionLevel string
	isEstimate       bool
	isVerbose        bool
	outFile          string
	disableHeadless  bool

	rootCmd = &cobra.Command{
		Use:   "pdfren",
		Short: "pdfren compresses your PDF using Adobe's online PDF compressor tool",
		Long:  "pdfren compresses your PDF using Adobe's online PDF compressor tool",
		Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

			if isVerbose {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			} else {
				zerolog.SetGlobalLevel(zerolog.InfoLevel)
			}

			pdfPath := args[0]
			f, err := os.Open(pdfPath)
			if err != nil {
				log.Fatal().Msg(err.Error())
			}

			RunCompressor(f, outFile)
		},
	}
)

func Execute() {
	rootCmd.PersistentFlags().StringVar(&compressionLevel, "compression", "high", "set the compression level \"high|medium|low\"")
	rootCmd.PersistentFlags().BoolVar(&isEstimate, "estimate", false, "when enabled, return estimated file saving when compressing")
	rootCmd.PersistentFlags().BoolVar(&isVerbose, "verbose", false, "verbose output mode")
	rootCmd.PersistentFlags().StringVar(&outFile, "outFile", "out.pdf", "specify the compressed PDF file path")
	rootCmd.PersistentFlags().BoolVar(&disableHeadless, "disableHeadless", false, "set true to disable headless mode")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func RunCompressor(file *os.File, outPath string) {
	if compressionLevel != "high" && compressionLevel != "medium" && compressionLevel != "low" {
		log.Fatal().Msg("--compression must be one of: high|medium|low")
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", !disableHeadless),
		chromedp.WindowSize(1920, 600),
		chromedp.DisableGPU,
		chromedp.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.50 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(
		allocCtx,
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	// timeout after 60 seconds
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	compressorURL := "https://www.adobe.com/acrobat/online/compress-pdf.html"
	submitBtn := `button[data-test-id="ls-footer-primary-compress-button"]`
	downloadBtn := `button[data-testid="lifecycle-complete-5-download-button"]`
	compressionBtn := fmt.Sprintf("input[data-test-id=\"compress-radio-option-%s\"]", compressionLevel)

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	done := make(chan string, 1)
	chromedp.ListenTarget(ctx, func(v interface{}) {
		if ev, ok := v.(*browser.EventDownloadProgress); ok {
			// log.Printf("state: %s, completed: %s\n", ev.State.String(), completed)
			if ev.State == browser.DownloadProgressStateCompleted {
				done <- ev.GUID
				close(done)
			}
		}
	})

	log.Debug().Msgf("navigating to %s", compressorURL)
	if err := chromedp.Run(ctx,
		chromedp.Navigate(compressorURL),
	); err != nil {
		log.Fatal().Msg(err.Error())
	}

	log.Debug().Msg("waiting for document and input to load")
	if err := chromedp.Run(ctx,
		//chromedp.WaitVisible(`body > footer`),
		chromedp.WaitEnabled(`input[accept=".pdf"]`, chromedp.NodeReady),
	); err != nil {
		log.Fatal().Msg(err.Error())
	}
	log.Debug().Msg("document and input ready")

	log.Debug().Msg("uploading pdf file")
	if err := chromedp.Run(ctx,
		chromedp.SetUploadFiles(`input[accept=".pdf"]`, []string{file.Name()}, chromedp.NodeVisible),
		chromedp.WaitVisible(`div[aria-label="Select compression level:"]`, chromedp.NodeVisible),
	); err != nil {
		log.Fatal().Msg(err.Error())
	}
	log.Debug().Msg("PDF uploaded")

	log.Debug().Msgf("setting compression level to %s", compressionLevel)
	if err := chromedp.Run(ctx,
		chromedp.Click(compressionBtn, chromedp.NodeReady),
	); err != nil {
		log.Fatal().Msg(err.Error())
	}
	log.Debug().Msg("compression set")

	log.Debug().Msg("running compressor")
	if err := chromedp.Run(ctx,
		chromedp.WaitEnabled(submitBtn, chromedp.NodeEnabled),
		chromedp.Click(submitBtn, chromedp.NodeEnabled),
		chromedp.WaitVisible(downloadBtn, chromedp.NodeVisible),
	); err != nil {
		log.Fatal().Msg(err.Error())
	}
	log.Debug().Msg("compression set")

	log.Debug().Msg("downloading compressed PDF")
	if err := chromedp.Run(ctx,
		chromedp.WaitVisible(downloadBtn, chromedp.NodeVisible),
		chromedp.WaitEnabled(downloadBtn, chromedp.NodeEnabled),
		//chromedp.Sleep(time.Second*1),
		browser.
			SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorAllowAndName).
			WithDownloadPath(wd).
			WithEventsEnabled(true),
		chromedp.Click(downloadBtn, chromedp.NodeReady),
	); err != nil {
		log.Fatal().Msg(err.Error())
	}
	log.Debug().Msg("downloaded compressed PDF")

	guid := <-done

	dlFile := filepath.Join(wd, guid)
	os.Rename(dlFile, outFile)
	log.Debug().Msgf("wrote %s", outFile)
}
