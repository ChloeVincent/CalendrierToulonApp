package main

import (
        "fmt"
        "log"
        "net/http"
        "context"
        "math/rand"
        "strconv"


        "google.golang.org/appengine"
        "golang.org/x/oauth2"
        "golang.org/x/oauth2/google"
        "google.golang.org/api/calendar/v3"
)

var (
    tokenCookies map[string]*oauth2.Token;
    stateString string;
)

func getStateString(updateString bool) string{
    if updateString {
        stateString = strconv.FormatInt(rand.Int63(),10)
    }
    return stateString
}

func init(){
    tokenCookies = make(map[string]*oauth2.Token)
}


// initializer function to avoid changes on what should be a constant
// (if it was possible in go)
func getConfig() *oauth2.Config{
    //redirectURL:= "https://indivision-toulon.ew.r.appspot.com/oauth2CallBack/"
    redirectURL := "http://localhost:8080/oauth2CallBack/"
    // note that the redirect url has to change depending on the environment (local test or appengine)

    //CLIENT_ID and CLIENT_SECRET are defined in another go file, ignored by git, to avoid committing it to GitHub
    // They can be retrieve from Google Cloud Platform > APIs & Services > Credentials > OAuth2 Clients
    return &oauth2.Config{
        ClientID:     CLIENT_ID,
        ClientSecret: CLIENT_SECRET,
        RedirectURL:  redirectURL,
        Scopes:       []string{"https://www.googleapis.com/auth/calendar.events.readonly"},
        Endpoint:     google.Endpoint,
    }
}

// Connect with OAuth2
func getLocalToken(w http.ResponseWriter, r *http.Request) *oauth2.Token {
    cookie, err := r.Cookie("CookieId")
    if err != nil{
        fmt.Println("Could not retrieve the cookie: "+err.Error())
        return nil
    } else{
        cookieId:= cookie.Value
        return tokenCookies[cookieId];
    }
}


func loginHandler(w http.ResponseWriter, r *http.Request) {
    if getLocalToken(w,r) != nil {
        fmt.Println("Already logged in : a cookie exists for this user")
        http.Redirect(w, r, "/", http.StatusSeeOther)
    } else {
        getOAuth2Link(w,r)
        return
    } 
}

func oauth2CallBackHandler(w http.ResponseWriter, r *http.Request) {
    updateTokenCookies(w,r)
    http.Redirect(w, r, "/", http.StatusSeeOther)
}


func getOAuth2Link(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Getting the OAuth2 token via link")

    authURL := getConfig().AuthCodeURL(getStateString(true), oauth2.AccessTypeOffline) 
    
    w.Header().Set("Content-Type", "text/html; charset=utf-8")    
    fmt.Fprint(w, "Go to the following link and sign in to your account: </br><a href="+ authURL+">Click link </a>")
}


func updateTokenCookies(w http.ResponseWriter, r *http.Request){
    fmt.Println("get authcode")
    if r.FormValue("state") != getStateString(false) {
        log.Fatalf("The state is invalid, closing the session")
    }

    authCode := r.FormValue("code")
    if authCode == "" {
        fmt.Println("The request does not contain an authentication code, please log in")
        return
    }

    fmt.Println("About to exchange authorization code for token")

    ctx := appengine.NewContext(r)

    cookieId := strconv.FormatInt(rand.Int63(),10)
    cookie := http.Cookie{Name : "CookieId", 
                           Value : cookieId,
                           Path: "/"}

    http.SetCookie(w, &cookie)
    var err error    
    tokenCookies[cookieId], err = getConfig().Exchange(ctx, authCode)
    if err != nil {
        fmt.Println("Error with authorization code exchange : " +err.Error())
        fmt.Println("\nAuthorization code is : "+authCode)
        fmt.Println("Please reload the page and follow the link provided")
        return
    }

    fmt.Println("token exchanged")
}

func startCalendarService(token *oauth2.Token) *calendar.Service{
    fmt.Println("Start Calendar Service")

    client:= getConfig().Client(context.Background(), token)
    fmt.Println("Client created")

    calendarService, err := calendar.New(client)
    if err != nil {
        log.Fatalf("Unable to retrieve Calendar client: %v", err)
    }
    fmt.Println("Calendar service started")
    return calendarService
}

