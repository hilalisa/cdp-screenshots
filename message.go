package main

type Message struct {
	HTML         string  `json:"html"` // either HTML or URL, preferred HTML
	URL          string  `json:"url"`
	Width        int64   `json:"width"`
	Height       int64   `json:"height"`
	Scaling      float64 `json:"scaling"`   // 1.00 by default
	Delay        int64   `json:"delay"`     // in ms
	FullPage     bool    `json:"full_page"` // take a screenshot of the full page
	Format       string  `json:"format"`    // jpeg or png
	Quality      int64   `json:"quality"`
	Callback     string  `json:"callback"`      // url of the callback
	CallbackType string  `json:"callback_type"` // "blob" or "s3", "blob" by default
}

func mustFloat64(val float64, err error) float64 {
	if err != nil {
		panic(err)
	}
	return val
}

func mustInt64(val int64, err error) int64 {
	if err != nil {
		panic(err)
	}
	return val
}
