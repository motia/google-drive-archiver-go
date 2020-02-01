package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

type myFile struct {
	id   string
	path string
}

func loadDirContents(
	srv *drive.Service,
	d *myFile,
	dirChan chan<- *myFile,
	fileChan chan<- *myFile,
	// dirWg *sync.WaitGroup,
) {
	for {
		q := fmt.Sprintf("'%s' in parents", d.id)
		r, err := srv.Files.List().
			Q(q).
			PageSize(10).
			Fields("nextPageToken, files(id, name, mimeType)").Do()
		if err != nil {
			log.Fatalf("Unable to retrieve files: %v", err)
			// log.Fatalf(err.Error())
		}

		for _, f := range r.Files {
			newFile := &myFile{
				f.Id,
				d.path + "/" + f.Name,
			}
			if f.MimeType == "application/vnd.google-apps.folder" {
				// fmt.Println("Adding directory: ", newFile)
				dirChan <- newFile
			} else {
				// fmt.Println("Adding File: ", newFile)
				fileChan <- newFile
			}
		}

		if r.NextPageToken == "" {
			break
		}
	}

	if len(dirChan) == 0 {
		close(dirChan)
	}
}

func onFileExplored(
	srv *drive.Service,
	d *myFile,
	fileWg *sync.WaitGroup,
) {
	defer fileWg.Done()

	// process file here
}

func main() {
	fmt.Println("Hello!")
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(
		b,
		drive.DriveMetadataReadonlyScope,
		drive.DriveReadonlyScope,
	)

	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := drive.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	dirChan := make(chan *myFile, 5)
	fileChannel := make(chan *myFile, 5)

	fmt.Println("Channels created!")

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		fmt.Println("Started loading directories")

		for d := range dirChan {
			fmt.Println("Directory found: ", d)
			loadDirContents(srv, d, dirChan, fileChannel)
		}
		close(fileChannel)

		wg.Done()
	}()

	dirChan <- &myFile{
		id:   "1TWHmapjJP0NfMMmCwjK3qNMjws1mncOn",
		path: "",
	}

	fmt.Println("Started loading files")

	for file := range fileChannel {
		wg.Add(1)
		fmt.Println("File found: ", file)
		go onFileExplored(srv, file, &wg)
	}

	fmt.Println("Waiting...........")

	wg.Wait()
	fmt.Println("Job Done!")
}
