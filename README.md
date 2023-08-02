# pdfren

## About

pdfren is a simple PDF compression tool that leverages Adobe's online PDF compression tool via automated browser via the chromedp package.

## Why?

The tools I found on Github rely on running ghostscript via an exec call. I think its better to have everything bundled.

Adobe's compression options out of the box seem to ensure maximally readable PDFs, so this tool only allows compression within those bounds. Ghostscript-reliant programs seem to give too much freedom, sometimes resulting in over-compressed, blurry outputs. 

## Running

Running pdfren looks like:
```bash
pdfren /path/to/my/file.pdf --compression medium --outFile compressed-file.pdf
```

See all flag options with the help flag:
```bash
pdfren --help
```

## Issues
I've noticed occasional timeouts, so you may have to re-run the command on occasion. This could be due to flaws in my JS selectors, changes to the adobe compression site, or some other issue.

## Contributions
If you like the tool and want to improve it, please do! I enjoy writing little scraper tools like this and would be thrilled to hear your suggestions.
