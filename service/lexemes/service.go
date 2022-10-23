package lexemes

import "github.com/blevesearch/segment"

type (

	// Service interface represents the text segmentation service.
	// http://www.unicode.org/reports/tr29/
	Service interface {

		// Split splits the input text to the slice of meaningful lexemes (excluding whitespaces, separators, etc).
		Split(input string) (lexemes []string)
	}

	service struct {
	}
)

func NewService() Service {
	return service{}
}

func (svc service) Split(input string) (lexemes []string) {
	segmenter := segment.NewWordSegmenterDirect([]byte(input))
	for eof := !segmenter.Segment(); !eof; eof = !segmenter.Segment() {
		if segmenter.Type() != segment.None {
			lexemes = append(lexemes, segmenter.Text())
		}
	}
	return
}
