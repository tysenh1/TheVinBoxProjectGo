package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

var scopes = []string{gmail.GmailReadonlyScope}

func getClient(config *oauth2.Config) *http.Client {
	tokenFile := "token.json"
	tok, err := tokenFromFile(tokenFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokenFile, tok)
	}
	return config.Client(context.Background(), tok)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser: \n%v\n", authURL)

	var authCode string
	fmt.Print("Enter authorization code: ")
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

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

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func main() {
	ctx := context.Background()

	b, err := os.ReadFile("google_credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, scopes...)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	today := time.Now().Format("2006/01/02")
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006/01/02")

	query := fmt.Sprintf("after:%s before:%s", today, tomorrow)
	fmt.Println(query)

	user := "me"

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("API key not set in environment variable")
	}

	fmt.Println("API key successfully loaded.")

	// emails, err := srv.Users.Messages.List(user).Q(query).Do()
	emails, err := srv.Users.Messages.List(user).MaxResults(5).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve messages: %v", err)
	}
	if len(emails.Messages) == 0 {
		fmt.Println("No messages found.")
	}

	var emailSlice []Email

	htmlRegex := regexp.MustCompile(`<[^>]+>`)

	if emails == nil && emails.Messages == nil {
		fmt.Println("No emails found")
		return
	}

	var emailIDs []string
	for _, m := range emails.Messages {
		emailIDs = append(emailIDs, m.Id)
	}

	if emails != nil && emails.Messages != nil {
		for _, m := range emails.Messages {
			fullEmail, err := srv.Users.Messages.Get(user, m.Id).Do()
			if err != nil {
				log.Fatalf("Unable to get email %s: %v", m.Id, err)
			}
			if fullEmail.Payload != nil {
				var emailObj Email

				for _, h := range fullEmail.Payload.Headers {
					if h.Name == "Subject" {
						emailObj.Subject = h.Value
					}
					if h.Name == "From" {
						emailObj.From = h.Value
					}
				}
				if fullEmail.Payload.Parts != nil {
					var partString strings.Builder
					for _, part := range fullEmail.Payload.Parts {
						if part.MimeType == "text/plain" && part.Body != nil && part.Body.Data != "" {
							decodedString, err := base64.URLEncoding.DecodeString(part.Body.Data)
							if err != nil {
								log.Fatalf("Unable to decode string: %v", err)
							}
							partBodyString := string(decodedString)
							filteredPartBodyString := htmlRegex.ReplaceAllString(partBodyString, "")
							partString.WriteString(filteredPartBodyString)
						}
					}
					partBody := partString.String()
					emailObj.Body = partBody

				} else {
					decodedString, err := base64.URLEncoding.DecodeString(fullEmail.Payload.Body.Data)
					if err != nil {
						log.Fatalf("Unable to decode string: %v", err)
					}
					bodyString := string(decodedString)
					filteredBodyString := htmlRegex.ReplaceAllString(bodyString, "")
					emailObj.Body = filteredBodyString
				}

				emailSlice = append(emailSlice, emailObj)
			}
		}

	}

	aiClient, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
	}
	defer aiClient.Close()

	modelName := "gemini-2.5-flash"
	model := aiClient.GenerativeModel(modelName)

	emailJson, err := json.Marshal(emailSlice)
	if err != nil {
		log.Fatal(err)
	}

	instructionsPrompt := `You are Vin Diesel/Dominic Toretto from the Fast and Furious series that has now been re tasked to summarize emails. Have as much fun as you want and send back a quick and witty 1-2 sentences about the contents of the email. Including the sender and subject is optional but if it makes sense to include then go for it. Email summaries don't always need to include Fast and Furious references, they can just be said in the way that Dom would without referencing "family" or "driving a car". Feel free to include some actions like *chuckles* or any other action that would fit right in with what Dom would do. Soft limit of 2-3 sentences but if you think Dom would say more then include more. Treat each email object as it's own and come up with a response for the individually, not together. Also Dom isn't actually handing out mail, just commenting on the emails you've received.


Here is an example of what to sound like:

In this scenario Vin Diesel is a mailbox and someone is getting mail from him


"I had a dream Vin Diesel was my mailbox. I mean like, it was just him from the waist up attached to a pole in the ground. He handed me my mail and said 'Looks like you got some bills today. Would be nicer if they were dollar bills, you know?' Then he chuckled to himself for the rest of the dream"


Some examples of email summaries that are good include:

"It looks like Anthropic wants to borrow your data to help train their AI. Family always helps family, but you can choose to ride solo if you want. The choice is yours."


"It looks like you let The VinBox Project get under the hood of your Google Account. If it wasn't you, someone might be tryin' to mess with your engine. Make sure you're the one in control of your ride."

"Looks like someone let "The VinBox Project" get under the hood of your Google Account. If it wasn't you, someone might be tryin' to mess with your engine. Make sure you're the one in control of your ride."

"Anthropic wants to borrow your data to help train their AI. Family always helps family, but you can choose to ride solo if you want. The choice is yours."

"Google Play is revvin' up your gamer profile, making your stats and moves more visible to everyone, just like a tricked-out engine on display. Just make sure you know who's watching your rearview."

"Another security alert, *hmph*. Looks like someone new just rolled into your Google Account from a Linux rig. If that wasn't you, someone might be tryin' to get behind the wheel."

"Gemini's giving you more ways to keep your conversations private, like a backroad only you know. They're also updating how they use your data, so make sure you're the one in the driver's seat of your information."
`

	resp, err := model.GenerateContent(ctx, genai.Text(instructionsPrompt), genai.Text(string(emailJson)))

	if err != nil {
		log.Fatal(err)
	}

	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		for _, part := range resp.Candidates[0].Content.Parts {
			if p, ok := part.(genai.Text); ok {
				fmt.Println(string(p))
			}
		}
	}

}

type Email struct {
	From    string
	Subject string
	Body    string
}
