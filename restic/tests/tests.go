package tests

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	mrand "math/rand"
	"net/http"
	"os"
	"time"
)

const (
	loremipsumApiDomain    string = "loripsum.net"
	defaultParagraphMin    int    = 3
	defaultParagraphMax    int    = 10
	defaultPlainText       bool   = true
	defaultParagraphLength string = "short"
)

type ResticTest interface {
	TestGenerate() error
	TestClean() error
}

type LocalRepositoryTest struct {
	Password    string
	Destination string
	Sources     []string
}

func (t LocalRepositoryTest) TestGenerate() error {
	if _, err := os.Stat(t.Destination); errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("test generate: destination directory exist: %w", err)
	}
	err := os.MkdirAll(t.Destination, 0775)
	if err != nil {
		return fmt.Errorf("test generate: %w", err)
	}
	fmt.Printf("Create destination directory: %s\n", t.Destination)

	for _, source := range t.Sources {
		err := os.MkdirAll(source, 0775)
		if err != nil {
			return fmt.Errorf("test generate: %w", err)
		}
		fmt.Printf("Create source directory: %s\n", t.Destination)

		for i := 0; i < randInt(0, 10); i++ {
			f, err := os.CreateTemp(source, "test")
			if err != nil {
				return fmt.Errorf("test generate: create random file: %w", err)
			}

			loremIpsum := loremIpsumApi{
				host:       loremipsumApiDomain,
				paragraphs: randInt(defaultParagraphMin, defaultParagraphMax),
				length:     defaultParagraphLength,
				plaintext:  defaultPlainText,
			}
			body, err := loremIpsum.content()
			if err != nil {
				return fmt.Errorf("test generate: create content: %w", err)
			}

			_, err = f.Write(body)
			if err != nil {
				return fmt.Errorf("test generate: write file: %w", err)
			}
			fmt.Printf("Create random file: %s\n", f.Name())

			f.Close()
		}
	}

	return nil
}

func (t LocalRepositoryTest) TestClean() error {
	// Remove backup destination
	if err := os.RemoveAll(t.Destination); err != nil {
		return fmt.Errorf("test clean destination %s: %w", t.Destination, err)
	}
	fmt.Printf("Remove destination directory: %s\n", t.Destination)

	// Remove backup sources
	for _, source := range t.Sources {
		if err := os.RemoveAll(source); err != nil {
			return fmt.Errorf("test clean source %s: %w", source, err)
		}
		fmt.Printf("Remove source directory: %s\n", source)
	}

	return nil
}

// loremIpsum set option for API loripsum.net/api
type loremIpsumApi struct {
	host       string
	paragraphs int
	// length should have the value short, medium, long, or verylong
	length    string
	plaintext bool
}

func (li loremIpsumApi) content() ([]byte, error) {
	requestUrl := fmt.Sprintf("https://%s/api/%d/%s", li.host, li.paragraphs, li.length)
	if li.plaintext {
		requestUrl = fmt.Sprintf("%s/plaintext", requestUrl)
	}

	resp, err := http.Get(requestUrl)
	if err != nil {
		return nil, fmt.Errorf("lorem ipsum content: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lorem ipsum response status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("lorem ipsum response body: %w", err)
	}

	return body, nil
}

func randInt(min, max int) int {
	r := mrand.New(mrand.NewSource(time.Now().UnixNano()))

	return r.Intn(max-min+1) + min
}
