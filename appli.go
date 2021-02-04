package main

import (
        "encoding/json"
        "fmt"
        "io/ioutil"
        "log"
        "net/http"
        "os"
        "time"
        "html/template"
        "strings"
        "strconv"

        "golang.org/x/net/context"
        "golang.org/x/oauth2"
        "golang.org/x/oauth2/google"
        "google.golang.org/api/calendar/v3"
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
                log.Fatalf("Unable to read authorization code: %v", err)
        }

        tok, err := config.Exchange(context.TODO(), authCode)
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

type Month struct {  
    Name string
    Index string
    Days  int
    StartDay int
    //OccupiedDays []int
}

func contains(s []OccupiedDay, e int) bool {
    for _, a := range s {
        //fmt.Printf("a: %v ; e: %v", a, e)
        if a.DayNumber == e {
            return true
        }
    }
    return false
}

func getColor(s []OccupiedDay, e int) string {
    for _, a := range s {
        //fmt.Printf("a: %v ; e: %v", a, e)
        if a.DayNumber == e {
            return a.Color
        }
    }
    return ""
}


// Runs server
func handler(w http.ResponseWriter, r *http.Request) {
    //fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
    //fmt.Fprintf(w, "<h1>Agenda Toulon</h1>")
    jan:= Month{"jan","01",31, 4}
    feb:= Month{"feb","02", 28, 0}
    mar:= Month{"mar","03",31, 0}
    avr:= Month{"avr","04", 30,3}
    may:= Month{"may","05", 31,5}
    jun:= Month{"jun","06", 30,1}
    jul:= Month{"jul","07", 31,3}
    aug:= Month{"aug","08", 31,6}
    sep:= Month{"sep","09", 30,2}
    oct:= Month{"oct","10", 31,4}
    nov:= Month{"nov","11", 30,0}
    dec:= Month{"dec","12", 31,2}

    months := []Month{jan,feb,mar, avr, may, jun, jul, aug, sep, oct, nov, dec}
    fmap:= template.FuncMap{
        "Iterate": func(count int) []int {
            var i int
            var Items []int
            for i = 0; i < (count); i++ {
                Items = append(Items, i+1)
            }
            return Items
        },
    }
    fmap2:= template.FuncMap{
        "IsOccupied": func(day int, monthStr string) bool {
            var month, _= strconv.Atoi(monthStr)
            return contains(OccupiedDaysList[month], day)
        },
    }
    fmap3:= template.FuncMap{
        "DayColor": func(day int, monthStr string) string {
            var month, _= strconv.Atoi(monthStr)
            return getColor(OccupiedDaysList[month], day)
        },
    }
    t, _ := template.New("template.html").Funcs(fmap).Funcs(fmap2).Funcs(fmap3).ParseFiles("template.html")

    t.Execute(w, months)
}



type OccupiedDay struct {  
    DayNumber int
    Color string
}

var OccupiedDaysList = make(map[int][]OccupiedDay)



func main() {
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

        srv, err := calendar.New(client)
        if err != nil {
                log.Fatalf("Unable to retrieve Calendar client: %v", err)
        }

        t := time.Now().Format(time.RFC3339)
        events, err := srv.Events.List("primary").ShowDeleted(false).
                SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
        if err != nil {
                log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
        }
        fmt.Println("Upcoming events:")
        if len(events.Items) == 0 {
                fmt.Println("No upcoming events found.")
        } else {
                

                for _, item := range events.Items {
                        startDate := item.Start.Date
                        colorID := item.ColorId
                        colorIdDict := map[string]string{
                                "11": "red",
                                "6": "orange",
                                "":"blue",
                                "1":"blue",
                                "2":"blue",
                                "3":"blue",
                                "4":"blue",
                                "5":"blue",
                                "7":"blue",
                                "8":"blue",
                                "9":"blue",
                                "10":"blue",
                            }

                        fmt.Printf("\ncolorID: %v aka %v \n",colorID, colorIdDict[colorID])
                        if startDate == "" {
                                fmt.Printf("Should add all-day events in calendar, on %v ", item.Start.DateTime)
                        }
                        fmt.Printf("%v (%v)\n", item.Summary, startDate)

                        var ymd = strings.Split(startDate, "-") 
                        m,_ := strconv.Atoi(ymd[1])
                        d,_ := strconv.Atoi(ymd[2])
                        fmt.Printf("year: %v, month: %v, day: %v ", ymd[0], m, d)
                        
                        OccupiedDaysList[m]= append(OccupiedDaysList[m], OccupiedDay{d, colorIdDict[colorID]})

                        endDate :=item.End.Date
                        if endDate != startDate {
                            var endYmd = strings.Split(endDate, "-") 
                            endM,_ := strconv.Atoi(endYmd[1])
                            endD,_ := strconv.Atoi(endYmd[2])
                            fmt.Printf("ENDDATE: year: %v, month: %v, day: %v ", endYmd[0], endM, endD)
                            
                            if m == endM{
                                for day := d; day < endD; day ++{
                                    OccupiedDaysList[m]= append(OccupiedDaysList[m], OccupiedDay{day, colorIdDict[colorID]})
                                }
                            } else{ // note that this does not work if stay lasts more than 2 months
                                for day := d; day < 32; day ++{
                                    OccupiedDaysList[m]= append(OccupiedDaysList[m], OccupiedDay{day, colorIdDict[colorID]})
                                }
                                for day := 0; day < endD; day ++{
                                    OccupiedDaysList[endM]= append(OccupiedDaysList[endM], OccupiedDay{day, colorIdDict[colorID]})
                                }
                            }    
                            

                        }



                        //fmt.Printf(Time.Month)
                }
        //fmt.Println()
        //fmt.Println(OccupiedDaysList)

        http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources")))) 
        http.HandleFunc("/", handler)
        log.Fatal(http.ListenAndServe(":8080", nil))
        }
}
