package main

import (
        "encoding/json"
        "fmt"
        "io/ioutil"
        "log"
        "net/http"
        "os"

        "golang.org/x/oauth2"
        "golang.org/x/oauth2/google"
        "google.golang.org/api/calendar/v3"
)

const isLocal = true;

func getOAuth2Link(w http.ResponseWriter, r *http.Request) {
    //this function is needed to build
}

// Calendar service functions: getClient, getTokenFromWeb, tokenFromFile, saveToken and startCalendarService
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
    return config.Client(ctx, tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
    authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
    fmt.Printf("Go to the following link in your browser then type the "+
        "authorization code: \n%v\n\nYou probably need to comment some lines in the main() ", authURL)

    var authCode string
    if _, err := fmt.Scan(&authCode); err != nil {
        log.Fatalf("Unable to read authorization code: %v", err)
    }
    log.Printf("about to exchange")
    tok, err := config.Exchange(ctx, authCode)
    if err != nil {
        log.Fatalf("Unable to retrieve token from web: %v", err)
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



func startCalendarService(w http.ResponseWriter,r *http.Request){

    b, err := ioutil.ReadFile("credentials.json")
    if err != nil {
            log.Fatalf("Unable to read client secret file: %v", err)
    }

    // If modifying these scopes, delete your previously saved token.json.
    config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
    if err != nil {
            log.Fatalf("Unable to parse client secret file to config: %v", err)
    }

    client := getClient(config)

    calendarService, err = calendar.New(client)
    if err != nil {
            log.Fatalf("Unable to retrieve Calendar client: %v", err)
    }

    tokenRequested = true
    calendarServiceStarted = true
    http.Redirect(w, r, "/", http.StatusSeeOther)
}