package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "context"

    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "google.golang.org/api/calendar/v3"
)

// initializer function to avoid changes on what should be a constant
// (if it was possible in go)
func getConfig() *oauth2.Config{
    b, err := ioutil.ReadFile("credentials.json")
    if err != nil {
        log.Fatalf("Unable to read client secret file: %v", err)
    }

    // If modifying these scopes, delete your previously saved token.json.
    oauth2Config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
    if err != nil {
        log.Fatalf("Unable to parse client secret file to config: %v", err)
    }
    return oauth2Config
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

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb() *oauth2.Token {
    oauth2Config := getConfig()
    authURL := oauth2Config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
    fmt.Printf("Go to the following link in your browser then type the "+
        "authorization code: \n%v\n\nNote that you probably need to comment some lines in the main()\n\n"+
        "Authorization code:", authURL)

    var authCode string
    if _, err := fmt.Scan(&authCode); err != nil {
        fmt.Printf("\nUnable to read authorization code: %v\n", err)
        return nil
    }
    log.Printf("about to exchange")
    tok, err := oauth2Config.Exchange(context.Background(), authCode)
    if err != nil {
        fmt.Printf("\nUnable to retrieve token from web: %v\n", err)
        return nil
    }
    return tok
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

func getLocalToken(w http.ResponseWriter, r *http.Request) *oauth2.Token {
    // The file token.json stores the user's access and refresh tokens, and is
    // created automatically when the authorization flow completes for the first
    // time.
    tokFile := "token.json"
    tok, err := tokenFromFile(tokFile)
    if err != nil {
            tok = getTokenFromWeb()
            if tok != nil{
                saveToken(tokFile, tok)
            }
    }
    return tok
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
    if getLocalToken(w,r) != nil {
        fmt.Println("Already logged in : a cookie exists for this user")
        http.Redirect(w, r, "/", http.StatusSeeOther)
    } else {
        fmt.Fprint(w, "There was an error, could not get token")
        return
    }
}

func oauth2CallBackHandler(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/", http.StatusSeeOther)
}

func startCalendarService(w http.ResponseWriter, token *oauth2.Token) *calendar.Service {
    client := getConfig().Client(context.Background(), token)

    calendarService, err := calendar.New(client)
    if err != nil {
            log.Fatalf("Unable to retrieve Calendar client: %v", err)
    }
    return calendarService
}

// delete cookies handler does nothing in this case
func deleteCookiesHandler(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/", http.StatusSeeOther)
}
