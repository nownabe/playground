package main

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sync"

	"cloud.google.com/go/storage"

	"github.com/labstack/echo"
)

const bucket = "YOUR_BUCKET"

func main() {
	e := echo.New()

	e.Static("/", "public")
	e.POST("/upload1", upload1)
	e.POST("/upload2", upload2)
	e.POST("/upload3", upload3)

	e.Start(":1323")
}

func upload1(c echo.Context) error {
	name := c.FormValue("name")

	fileHeader, err := c.FormFile("file")
	if err != nil {
		fmt.Println(err)
		return err
	}

	file, err := fileHeader.Open()
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer file.Close()

	var filename string
	switch fileHeader.Header["Content-Type"][0] {
	case "image/png":
		filename = name + ".png"
	case "image/jpeg":
		filename = name + ".jpg"
	default:
		return fmt.Errorf("Unknown filetype")
	}

	if err := putGCS(filename, file); err != nil {
		fmt.Println(err)
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{
		"success": "true",
		"name":    name,
		"link":    "//storage.googleapis.com/" + bucket + "/" + filename,
	})
}

func upload2(c echo.Context) error {
	name := c.FormValue("name")

	form, err := c.MultipartForm()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fileHeaders := form.File["files"]
	links := make([]string, len(fileHeaders), len(fileHeaders))

	for i, fileHeader := range fileHeaders {
		file, err := fileHeader.Open()
		if err != nil {
			fmt.Println(err)
			return err
		}
		defer file.Close()

		var ext string
		switch fileHeader.Header["Content-Type"][0] {
		case "image/png":
			ext = ".png"
		case "image/jpeg":
			ext = ".jpg"
		default:
			return fmt.Errorf("Unknown filetype")
		}
		filename := fmt.Sprintf("multi/%s%d%s", name, i, ext)

		if err := putGCS(filename, file); err != nil {
			fmt.Println(err)
			return err
		}

		links[i] = "//storage.googleapis.com/" + bucket + "/" + filename
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": "true",
		"name":    name,
		"links":   links,
	})
}

type task struct {
	I          int
	FileHeader *multipart.FileHeader
}

func upload3(c echo.Context) error {
	name := c.FormValue("name")

	var wg sync.WaitGroup
	file := make(chan *task)
	linkCh := make(chan string)
	linksCh := make(chan []string)

	form, err := c.MultipartForm()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fileHeaders := form.File["files"]

	// Link collector
	go func() {
		links := make([]string, len(fileHeaders), len(fileHeaders))
		for i := 0; ; i++ {
			l, ok := <-linkCh
			if ok {
				links[i] = l
			} else {
				break
			}
		}
		linksCh <- links
	}()

	// Worker
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				task, ok := <-file
				if !ok {
					return
				}
				i := task.I
				fileHeader := task.FileHeader
				fmt.Println("Working...", i, fileHeader.Filename)

				file, err := fileHeader.Open()
				if err != nil {
					fmt.Println(err)
					continue
				}
				defer file.Close()

				var ext string
				switch fileHeader.Header["Content-Type"][0] {
				case "image/png":
					ext = ".png"
				case "image/jpeg":
					ext = ".jpg"
				default:
					continue
				}
				filename := fmt.Sprintf("multi/%s%d%s", name, i, ext)

				if err := putGCS(filename, file); err != nil {
					fmt.Println(err)
					continue
				}

				linkCh <- "//storage.googleapis.com/" + bucket + "/" + filename
				// time.Sleep(1 * time.Second)
				fmt.Println("Finished work", i, fileHeader.Filename)
			}
		}()
	}

	for i, fileHeader := range fileHeaders {
		file <- &task{I: i, FileHeader: fileHeader}
		fmt.Println("Queued:", i, fileHeader.Filename)
	}
	close(file)

	wg.Wait()
	close(linkCh)

	links := <-linksCh

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": "true",
		"name":    name,
		"links":   links,
	})

}

func putGCS(name string, reader io.Reader) error {
	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer client.Close()

	bkt := client.Bucket(bucket)
	obj := bkt.Object(name)
	w := obj.NewWriter(ctx)

	if _, err := io.Copy(w, reader); err != nil {
		fmt.Println(err)
		return err
	}

	if err := w.Close(); err != nil {
		fmt.Println(err)
		return err
	}

	if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
