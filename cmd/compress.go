package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/chromedp"
	"github.com/spf13/cobra"
)

var (
	compressionLevel string
	isEstimate       bool

	rootCmd = &cobra.Command{
		Use:   "pdffren",
		Short: "pdffren compresses your PDF using Adobe's online PDF compressor tool",
		Long:  "pdffren compresses your PDF using Adobe's online PDF compressor tool",
		Run: func(cmd *cobra.Command, args []string) {
			RunScraper()
		},
	}
)

func Execute() {
	rootCmd.PersistentFlags().StringVar(&compressionLevel, "compression", "high", "set the compression level \"high|medium|low\"")
	rootCmd.PersistentFlags().BoolVar(&isEstimate, "estimate", false, "when enabled, return estimated file saving when compressing")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func RunScraper() {
	if compressionLevel != "high" && compressionLevel != "medium" && compressionLevel != "low" {
		log.Fatal("--compression must be one of: high|medium|low")
	}

	opts := make([]func(*chromedp.ExecAllocator), 0)
	chromedp.Flag("headless", false)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

	// create context
	ctx, cancel := chromedp.NewContext(
		allocCtx,
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	// create a timeout as a safety net to prevent any infinite wait loops
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// TODO: use argument/flg
	filename := "/Users/gb/Downloads/Invoice-C01BC5FE-0001.pdf"
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	submitBtn := `button[data-test-id="ls-footer-primary-compress-button"]`
	downloadBtn := `button[data-testid="lifecycle-complete-5-download-button"]`

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan string, 1)
	// set up a listener to watch the download events and close the channel
	// when complete this could be expanded to handle multiple downloads
	// through creating a guid map, monitor download urls via
	// EventDownloadWillBegin, etc
	chromedp.ListenTarget(ctx, func(v interface{}) {
		if ev, ok := v.(*browser.EventDownloadProgress); ok {
			// log.Printf("state: %s, completed: %s\n", ev.State.String(), completed)
			if ev.State == browser.DownloadProgressStateCompleted {
				done <- ev.GUID
				close(done)
			}
		}
	})

	if err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.adobe.com/acrobat/online/compress-pdf.html"),
		chromedp.WaitVisible(`body > footer`),
		chromedp.SetUploadFiles(`input[accept=".pdf"]`, []string{file.Name()}, chromedp.NodeVisible),
		chromedp.WaitVisible(`div[aria-label="Select compression level:"]`, chromedp.NodeVisible),
		chromedp.WaitEnabled(submitBtn, chromedp.NodeEnabled),
		chromedp.Click(submitBtn, chromedp.NodeEnabled),
		chromedp.WaitVisible(downloadBtn, chromedp.NodeVisible),
		chromedp.WaitEnabled(downloadBtn, chromedp.NodeEnabled),
		browser.
			SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorAllowAndName).
			WithDownloadPath(wd).
			WithEventsEnabled(true),
		chromedp.Click(downloadBtn, chromedp.NodeEnabled),
		//chromedp.ActionFunc(func(context.Context) error {
		//	fmt.Println("HERE")
		//	return nil
		//}),
	); err != nil {
		log.Fatal(err)
	}
	// This will block until the chromedp listener closes the channel
	guid := <-done

	// We can predict the exact file location and name here because of how we
	// configured SetDownloadBehavior and WithDownloadPath
	log.Printf("wrote %s", filepath.Join(wd, guid+".zip"))
	dlFile := filepath.Join(wd, guid)
	destFile := filepath.Join(wd, "test.pdf")
	os.Rename(dlFile, destFile)

	// time.Sleep(time.Second * 10000)
}
