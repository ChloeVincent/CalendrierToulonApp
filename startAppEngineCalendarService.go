package main

import (
        "fmt"
        "log"
        "net/http"
        "os"

        "golang.org/x/oauth2"
        "golang.org/x/oauth2/google"
        "google.golang.org/api/calendar/v3"
)

const isLocal = false;

func getOAuth2Link(w http.ResponseWriter, r *http.Request) {
    redirectURL := os.Getenv("OAUTH2_CALLBACK")
    if redirectURL == "" {
            //redirectURL = "https://indivision-toulon.ew.r.appspot.com/"
            redirectURL = "http://localhost:8080/"
            // note that the redirect url has to change depending on the environment (local test or appengine)
    }

    //CLIENT_ID and CLIENT_SECRET are defined in another go file, ignored by git, to avoid committing it to GitHub
    oauth2Config = &oauth2.Config{
        ClientID:     CLIENT_ID,
        ClientSecret: CLIENT_SECRET,
        RedirectURL:  redirectURL,
        Scopes:       []string{"https://www.googleapis.com/auth/calendar.events.readonly"},
        Endpoint:     google.Endpoint,
    }

    authURL := oauth2Config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

        
    fmt.Fprint(w, "Go to the following link and sign in to your account: \n", authURL)
    tokenRequested = true
}


func startCalendarService(w http.ResponseWriter, r *http.Request){
    fmt.Println("Start Calendar Service")

    if r.FormValue("code") == "" {
        fmt.Fprint(w, "Please reload the page and follow the link provided")
        tokenRequested = false
        return
    }

    var err error
    
    var tok *oauth2.Token
    fmt.Println("About to exchange authorization code for token")
    if ctx ==nil{
        fmt.Println("context is nil!")
    }

    authCode := r.FormValue("code")
    tok, err = oauth2Config.Exchange(ctx, authCode)
    if err != nil {
        fmt.Fprint(w, "Error with authorization code exchange : " +err.Error())
        fmt.Fprint(w, "\nAuthorization code is : "+authCode)
        fmt.Fprint(w, "Please reload the page and follow the link provided")
        tokenRequested = false
        return
    }
    
    fmt.Println("About to create client")
    client:= oauth2Config.Client(ctx, tok)
    fmt.Println("Client created")

    calendarService, err = calendar.New(client)
    if err != nil {
            log.Fatalf("Unable to retrieve Calendar client: %v", err)
            fmt.Fprint(w, "error getting calendar")
    }
     fmt.Println("Calendar service started")

    calendarServiceStarted = true
    http.Redirect(w, r, "/", http.StatusSeeOther)
}

