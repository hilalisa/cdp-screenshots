package screenshot

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/dom"
	"github.com/mafredri/cdp/protocol/emulation"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/nfnt/resize"
	"github.com/pkg/errors"
)

func TakeScreenshot(
	ctx context.Context, client *cdp.Client,
	url string, width, height int, scaling float64,
	delay time.Duration, fullPage bool, format string,
	quality int,
) ([]byte, error) {
	// Open a DOMContentEventFired client to buffer this event.
	domContent, err := client.Page.DOMContentEventFired(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to setup a listener to DOMContentEventFired")
	}
	defer domContent.Close()

	// Enable events on the Page domain, it's often preferrable to create
	// event clients before enabling events so that we don't miss any.
	if err = client.Page.Enable(ctx); err != nil {
		return nil, errors.Wrap(err, "")
	}
	if err = client.DOM.Enable(ctx); err != nil {
		return nil, errors.Wrap(err, "")
	}

	log.Print("Enabled Page and DOM events")

	// Prepare the viewport.
	if err := client.Emulation.SetDeviceMetricsOverride(ctx, &emulation.SetDeviceMetricsOverrideArgs{
		Width:             width,
		Height:            height,
		DeviceScaleFactor: 0,
		Mobile:            false,
	}); err != nil {
		return nil, errors.Wrap(err, "unable to set the initial viewport size")
	}
	if err := client.Emulation.SetVisibleSize(ctx, &emulation.SetVisibleSizeArgs{
		Width:  width,
		Height: height,
	}); err != nil {
		return nil, errors.Wrap(err, "unable to set the initial visible size")
	}

	log.Print("Set the page size")

	// Create the Navigate arguments with the optional Referrer field set.
	navArgs := page.NewNavigateArgs(url)
	if _, err := client.Page.Navigate(ctx, navArgs); err != nil {
		return nil, errors.Wrap(err, "unable to navigate to the page")
	}

	// Wait for the page to load
	if _, err = domContent.Recv(); err != nil {
		return nil, errors.Wrap(err, "unable to wait for the content to load")
	}

	log.Print("Navigated to the page")

	if delay != 0 {
		time.Sleep(delay)
	}

	if fullPage {
		// Fetch the document root node. We can pass nil here
		// since this method only takes optional arguments.
		doc, err := client.DOM.GetDocument(ctx, nil)
		if err != nil {
			return nil, errors.Wrap(err, "uanble to get the DOM document")
		}

		// Select the body element to figure out its height
		qsReply, err := client.DOM.QuerySelector(ctx, &dom.QuerySelectorArgs{
			Selector: "body",
			NodeID:   doc.Root.NodeID,
		})
		if err != nil {
			return nil, errors.Wrap(err, "unable to find the body element")
		}

		// Get body's size
		bmReply, err := client.DOM.GetBoxModel(ctx, &dom.GetBoxModelArgs{
			NodeID: &qsReply.NodeID,
		})
		if err != nil {
			return nil, errors.Wrap(err, "unable to get body's box model")
		}

		// And prepare the final viewport
		if err := client.Emulation.SetDeviceMetricsOverride(ctx, &emulation.SetDeviceMetricsOverrideArgs{
			Width:             width,
			Height:            bmReply.Model.Height,
			DeviceScaleFactor: 1,
			Mobile:            false,
		}); err != nil {
			return nil, errors.Wrap(err, "unable to set the final viewport size")
		}
		if err := client.Emulation.SetVisibleSize(ctx, &emulation.SetVisibleSizeArgs{
			Width:  width,
			Height: bmReply.Model.Height,
		}); err != nil {
			return nil, errors.Wrap(err, "unable to set the final visible size")
		}
	}

	// Capture a screenshot of the current page.
	screenshotArgs := page.NewCaptureScreenshotArgs()
	if format == "" {
		format = "png"
	}
	screenshotArgs = screenshotArgs.SetFormat(format)

	if format == "jpeg" && quality != 0 {
		screenshotArgs = screenshotArgs.SetQuality(quality)
	}

	if format != "png" && format != "jpeg" {
		return nil, errors.New("invalid format type")
	}

	log.Print("Starting capturing the screenshot")

	log.Printf("%+v", screenshotArgs)

	screenshot, err := client.Page.CaptureScreenshot(
		ctx, screenshotArgs,
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to take a screenshot of the page")
	}

	log.Print("Captured the screenshot")

	if scaling != 0 && scaling != 1 {
		var img image.Image
		if format == "png" {
			img, err = png.Decode(bytes.NewReader(screenshot.Data))
			if err != nil {
				return nil, errors.Wrap(err, "unable to parse the png screenshot")
			}
		} else if format == "jpeg" {
			img, err = jpeg.Decode(bytes.NewReader(screenshot.Data))
			if err != nil {
				return nil, errors.Wrap(err, "unable to parse the jpeg screenshot")
			}
		}

		resized := resize.Resize(
			uint(float64(width)*scaling),
			0,
			img,
			resize.Bicubic,
		)

		buf := &bytes.Buffer{}
		if format == "png" {
			if err := png.Encode(buf, resized); err != nil {
				return nil, errors.Wrap(err, "unable to encode the resized png image")
			}
		} else if format == "jpeg" {
			if err := jpeg.Encode(buf, resized, &jpeg.Options{
				Quality: quality,
			}); err != nil {
				return nil, errors.Wrap(err, "unable to encode the resized jpeg image")
			}
		}
		screenshot.Data = buf.Bytes()
	}

	return screenshot.Data, nil
}
