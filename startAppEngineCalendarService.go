package main

import (
    "fmt"
    "log"
    "net/http"
    "context"
    "math/rand"
    "strconv"
    "time"

    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "google.golang.org/api/calendar/v3"
)

var stateString string;

func getStateString(updateString bool) string{
    if updateString {
        stateString = strconv.FormatInt(rand.Int63(),10)
    }
    return stateString
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

// Login: get local token (=from cookie) or request an authentication link
func loginHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Println("loginHandler")

    if getLocalToken(w,r) != nil {
        fmt.Println("Already logged in : a cookie exists for this user")
        http.Redirect(w, r, "/", http.StatusSeeOther)

    } else {
        getOAuth2Link(w,r)
        return
    } 
}

// local token + get cookie
func getCookie(cookieName string , r *http.Request) (string, error) {
    cookie, err := r.Cookie(cookieName)
    if err != nil{
        //fmt.Println("Could not retrieve the cookie for "+cookieName+": "+err.Error())
        return "", err 
    } else{
        cookieValue:= cookie.Value
        //fmt.Println("A cookie exists, the token cookie is: '"+ cookieValue+"'")
        return cookieValue, nil;
    }
}

func getLocalToken(w http.ResponseWriter, r *http.Request) *oauth2.Token {
    var token  *oauth2.Token;

    accessToken, err1 := getCookie("AccessToken", r)
    refreshToken, err2 := getCookie("RefreshToken", r)
    tokenType, err3 := getCookie("TokenType", r)
    expiry, _:= getCookie("Expiry", r)
    if err1 ==nil && err2 == nil && err3 == nil {
        token = &oauth2.Token{AccessToken: accessToken, RefreshToken: refreshToken, TokenType: tokenType}
        expDate, err4 := time.Parse(time.UnixDate, expiry)
        if err4 != nil{
            fmt.Println("error parsing date: "+ err4.Error())
        } else{
            token.Expiry = expDate
        }
    }
    return token;

}

// get authentication link + call back handler that updates the cookies
func setCookie(w http.ResponseWriter, cookieName string, cookieValue string){
    expiration := time.Now().Add(30 * 24 * time.Hour)
    cookie := http.Cookie{Name : cookieName, 
                          Value : cookieValue,
                          Path: "/",
                          Expires: expiration}


    http.SetCookie(w, &cookie)
}

func saveTokenCookies(w http.ResponseWriter,token *oauth2.Token){
    setCookie(w, "AccessToken", token.AccessToken)
    setCookie(w, "RefreshToken", token.RefreshToken)
    setCookie(w, "TokenType", token.TokenType)
    setCookie(w, "Expiry", token.Expiry.Format(time.UnixDate))
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
    token, err := getConfig().Exchange(context.Background(), authCode, oauth2.AccessTypeOffline)
    if err != nil {
        fmt.Println("Error with authorization code exchange : " +err.Error())
        fmt.Println("\nAuthorization code is : "+authCode)
        fmt.Println("Please reload the page and follow the link provided")
        return
    }
    fmt.Println("Token exchanged")

    saveTokenCookies(w, token)
    fmt.Println("Cookies saved")
}

func oauth2CallBackHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Println("call back handler")
    updateTokenCookies(w,r)
    http.Redirect(w, r, "/", http.StatusSeeOther)
}

func getOAuth2Link(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Getting the OAuth2 token via link")
    authURL := getConfig().AuthCodeURL(getStateString(true), oauth2.AccessTypeOffline) 
    
    w.Header().Set("Content-Type", "text/html; charset=utf-8")    
    fmt.Fprint(w, "Go to the following link and sign in to your account: </br><a href="+ authURL+">Click link </a>")
}


// func checkAndUpdateToken(w http.ResponseWriter, token *oauth2.Token) *oauth2.Token {
//     tokenSource := getConfig().TokenSource(oauth2.NoContext, token)
//     newToken, err := tokenSource.Token()
//     fmt.Println(token.AccessToken)
//     fmt.Println(newToken.AccessToken)
//     fmt.Println(token.RefreshToken)
//     fmt.Println(newToken.RefreshToken)
//     fmt.Println(token.TokenType)
//     fmt.Println(newToken.TokenType)
//     fmt.Println(token.Expiry)
//     fmt.Println(newToken.Expiry)
//     if err != nil {
//         log.Fatalln(err)
//     }

//     if newToken.AccessToken != token.AccessToken {
//         saveTokenCookies(w, newToken)
//         log.Println("Saved new token:", newToken.AccessToken)
//         return newToken
//     }
//     return token
// }


// starts and returns calendar service with oauth2 token
func startCalendarService(w http.ResponseWriter, token *oauth2.Token) *calendar.Service{
    fmt.Println("Start Calendar Service")

    //token = checkAndUpdateToken(w, token)
    //should be updated by the following (?)
    client:= getConfig().Client(context.Background(), token)
    fmt.Println("Client created")

    calendarService, err := calendar.New(client)
    if err != nil {
        log.Fatalf("Unable to retrieve Calendar client: %v", err)
    }
    fmt.Println("Calendar service started")
    return calendarService
}

