package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/kylerky/everyphoto/labeler"
	gvision "google.golang.org/genproto/googleapis/cloud/vision/v1"
)

var concurrancy = 4

type labelAnno struct {
	entities []*gvision.EntityAnnotation
	path     string
}

type faceAnno struct {
	entities []*gvision.FaceAnnotation
	path     string
}

type webAnno struct {
	entity *gvision.WebDetection
	path   string
}

type annotations struct {
	Path   string    `json:"path"`
	Labels []string  `json:"labels"`
	Scores []float32 `json:"scores"`
}

func labelHandler(labels chan labelAnno, quit *sync.WaitGroup) {
	defer quit.Done()

	conn, err := net.Dial("tcp", os.Args[2])
	if err != nil {
		log.Fatal("label handler failed to connect to server:", err)
		return
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)

	for label := range labels {
		result := annotations{}
		result.Path = label.path

		for _, tag := range label.entities {
			// label amplifying

			/*
				Tag:   tag.Description,
				Score: tag.Score,
			*/
			result.Labels = append(result.Labels, tag.Description)
			result.Scores = append(result.Scores, tag.Score)
		}

		// talk to the server
		err := encoder.Encode(result)
		if err != nil {
			log.Fatal("Failed to encode the message and pass it to server:", err)
		}
	}
}

func faceHandler(faces chan faceAnno, quit *sync.WaitGroup) {
	defer quit.Done()

	conn, err := net.Dial("tcp", os.Args[2])
	if err != nil {
		log.Fatal("face handler failed to connect to server: ", err)
		return
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)

	for face := range faces {
		result := annotations{}
		result.Path = face.path

		for _, tag := range face.entities {
			// TODO
			// label amplifying
			if tag.JoyLikelihood == gvision.Likelihood_LIKELY || tag.JoyLikelihood == gvision.Likelihood_VERY_LIKELY {
				result.Labels = append(result.Labels, "joy")
				result.Scores = append(result.Scores, float32(tag.JoyLikelihood)/float32(gvision.Likelihood_VERY_LIKELY))
				result.Labels = append(result.Labels, "delighted")
				result.Scores = append(result.Scores, float32(tag.JoyLikelihood)/float32(gvision.Likelihood_VERY_LIKELY))
				result.Labels = append(result.Labels, "happy")
				result.Scores = append(result.Scores, float32(tag.JoyLikelihood)/float32(gvision.Likelihood_VERY_LIKELY))
			}
			if tag.SorrowLikelihood == gvision.Likelihood_LIKELY || tag.SorrowLikelihood == gvision.Likelihood_VERY_LIKELY {
				result.Labels = append(result.Labels, "sorrow")
				result.Scores = append(result.Scores, float32(tag.SorrowLikelihood)/float32(gvision.Likelihood_VERY_LIKELY))
				result.Labels = append(result.Labels, "sad")
				result.Scores = append(result.Scores, float32(tag.SorrowLikelihood)/float32(gvision.Likelihood_VERY_LIKELY))
				result.Labels = append(result.Labels, "painful")
				result.Scores = append(result.Scores, float32(tag.SorrowLikelihood)/float32(gvision.Likelihood_VERY_LIKELY))
			}
			if tag.AngerLikelihood == gvision.Likelihood_LIKELY || tag.AngerLikelihood == gvision.Likelihood_VERY_LIKELY {
				result.Labels = append(result.Labels, "anger")
				result.Scores = append(result.Scores, float32(tag.AngerLikelihood)/float32(gvision.Likelihood_VERY_LIKELY))
				result.Labels = append(result.Labels, "indignation")
				result.Scores = append(result.Scores, float32(tag.AngerLikelihood)/float32(gvision.Likelihood_VERY_LIKELY))
				result.Labels = append(result.Labels, "furious")
				result.Scores = append(result.Scores, float32(tag.AngerLikelihood)/float32(gvision.Likelihood_VERY_LIKELY))
			}
			if tag.SurpriseLikelihood == gvision.Likelihood_LIKELY || tag.SurpriseLikelihood == gvision.Likelihood_VERY_LIKELY {
				result.Labels = append(result.Labels, "surprise")
				result.Scores = append(result.Scores, float32(tag.SurpriseLikelihood)/float32(gvision.Likelihood_VERY_LIKELY))
				result.Labels = append(result.Labels, "astound")
				result.Scores = append(result.Scores, float32(tag.SurpriseLikelihood)/float32(gvision.Likelihood_VERY_LIKELY))
				result.Labels = append(result.Labels, "shock")
				result.Scores = append(result.Scores, float32(tag.SurpriseLikelihood)/float32(gvision.Likelihood_VERY_LIKELY))
			}
		}
		if len(result.Labels) > 0 {
			fmt.Println(result)
			err := encoder.Encode(result)
			if err != nil {
				log.Fatal("Failed to encode the message and pass it to server: ", err)
			}
		}
	}
}

/*
func webHandler(webs chan webAnno, quit *sync.WaitGroup) {
	defer quit.Done()

	conn, err := net.Dial("tcp", os.Args[2])
	if err != nil {
		log.Println("web handler failed to connect to server:", err)
		return
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	for web := range webs {
		result := annotations{}
		result.Path = web.path

		// TODO
		result := annotations{}
		result.Path = web.path

		for _, entity := range web.entity.WebEntities {
			result.Labels = append(result.Labels, entity.Description)
			result.Scores = append(result.Scores, entity.Score)
		}
		err := encoder.Encode(result)
		if err != nil {
			log.Println("Failed to encode the message and pass it to server:", err)
		}

	}
}
*/
func main() {

	if len(os.Args) != 3 {
		fmt.Println("usage: labeler <path> <server_address>")
		fmt.Println("example: labeler . localhost:8080")
		os.Exit(1)
	}

	pathCh := make(chan string, concurrancy)

	info, err := os.Stat(os.Args[1])
	if err != nil {
		log.Fatal("Cannot read file info: ", err)
	}

	if info.IsDir() {
		// start the traverser
		go labeler.Traverse(os.Args[1], pathCh)
	} else {
		concurrancy = 1
		pathCh <- os.Args[1]
		close(pathCh)
	}

	// channels for communicating labels
	labelCh := make(chan labelAnno)
	faceCh := make(chan faceAnno)
	//	webCh := make(chan webAnno)

	var handlers sync.WaitGroup
	handlers.Add(2)
	// go handle some label mappings
	// and also talk to the server
	go labelHandler(labelCh, &handlers)
	go faceHandler(faceCh, &handlers)
	//	go webHandler(webCh, &handlers)

	var labelers sync.WaitGroup
	labelers.Add(concurrancy)
	// start some (determined by concurrancy) goroutines
	for i := 0; i < concurrancy; i++ {
		go func(index int, lablers *sync.WaitGroup) {
			defer labelers.Done()
			for path := range pathCh {
				log.Println("routine", index, "running")
				// open file for reading
				file, err := os.Open(path)
				if err != nil {
					log.Println("Cannot open file:", err)
					continue
				}
				defer file.Close()

				// get file info
				stat, err := file.Stat()
				if err != nil {
					log.Println("Cannot read file status:", err)
					continue
				}

				log.Println("routine", index, "begins to compress")
				var buffer bytes.Buffer
				// compress the image if its size is larger than 2MB
				// directly copy to buffer otherwise
				if stat.Size() > 2<<20 {
					err := labeler.Compress(&buffer, file, 2<<20)
					if err != nil {
						log.Println("Cannot compress image:", err)
						continue
					}
				} else {
					_, err := buffer.ReadFrom(file)
					if err != nil {
						log.Println("Cannot read file to buffer:", err)
						continue
					}
				}
				log.Println("routine", index, "done with compression")

				// get labels
				log.Println("routine", index, "begins to label")
				labels, faces, _, err := labeler.Label(&buffer)
				if err != nil {
					log.Println("Cannot label the photo:", err)
					continue
				} else {
					labelCh <- labelAnno{entities: labels, path: path}
					faceCh <- faceAnno{entities: faces, path: path}
					//		webCh <- webAnno{entity: webs, path: path}
				}
				log.Println("routine", index, "done with labeling")
			}
		}(i, &labelers)
	}

	log.Println("waiting for labelers...")
	// wait for the labelers
	labelers.Wait()
	close(labelCh)
	close(faceCh)
	//	close(webCh)

	log.Println("waiting for handlers")
	// wait for the 3 handlers
	handlers.Wait()
}