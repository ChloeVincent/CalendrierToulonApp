package main

import (
        "fmt"
        "io"
        "log"
        "net/http"
        "os"
        "time"
        "html/template"
        "strings"
        "strconv"
        "bufio"
        "sync"
        "context"

        "golang.org/x/oauth2"
        "google.golang.org/api/calendar/v3"
)

type OccupiedDay struct {  
    DayNumber int
    Color string
}

type Month struct {  
    Name string
    Index string
    Days  int
    StartDay int
    //OccupiedDays []int
}

type TemplateArgs struct {
    Months []Month
    Events []string
}


var calendarService  *calendar.Service;
var ctx context.Context;

var MonthsDefinition = []Month {
    Month{"jan","01",31, 4},
    Month{"feb","02", 28, 0},
    Month{"mar","03",31, 0},
    Month{"avr","04", 30,3},
    Month{"may","05", 31,5},
    Month{"jun","06", 30,1},
    Month{"jul","07", 31,3},
    Month{"aug","08", 31,6},
    Month{"sep","09", 30,2},
    Month{"oct","10", 31,4},
    Month{"nov","11", 30,0},
    Month{"dec","12", 31,2}}

var colorIdDict = map[string]string{
    "11": "red",
    "6": "orange",
    "":"blue",
    "1":"blue",
    "2":"green",
    "3":"blue",
    "4":"blue",
    "5":"blue",
    "7":"blue",
    "8":"blue",
    "9":"blue",
    "10":"blue",
}


// Handler functions: iterate, contains and getColor
func iterate(count int) []int {
    var i int
    var Items []int
    for i = 0; i < (count); i++ {
        Items = append(Items, i+1)
    }
    return Items
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

var calendarServiceStarted bool;
var tokenRequested bool;

// Runs server
func handler(w http.ResponseWriter, r *http.Request) {

    if(!isLocal){
        if !tokenRequested{
            getOAuth2Link(w,r)
            return
        }
        if !calendarServiceStarted{
            startCalendarService(w,r)
            return
        }   
    } else{
        if !calendarServiceStarted {
            startCalendarService(w,r)
        }
    }

    if calendarService == nil{
        log.Fatal("Calendar service was not initialized properly.")
    }

    OccupiedDaysList := make(map[int][]OccupiedDay)
    events:= []string{"event1", "event2"}
    if (ctx != nil){
        OccupiedDaysList, events = refreshOccupiedDaysList()
    } else{
        log.Fatal("The background context was not initialized properly")
    }

    fmap:= template.FuncMap{"Iterate": iterate,
                            "IsOccupied": func(day int, monthStr string) bool {
                                        var month, _= strconv.Atoi(monthStr)
                                        return contains(OccupiedDaysList[month], day)},
                            "DayColor": func(day int, monthStr string) string {
                                        var month, _= strconv.Atoi(monthStr)
                                        return getColor(OccupiedDaysList[month], day)},
                            }
    
    t, _ := template.New("template.html").Funcs(fmap).ParseFiles("template.html")

    t.Execute(w, TemplateArgs{MonthsDefinition, events})
}


// refreshOccupiedDaysList functions: getYMD (Year Month Day) and appendODL (OccupiedDaysList)
func getYMD(startDate string)(int, int, int){
    ymd := strings.Split(startDate, "-")
    y,_ := strconv.Atoi(ymd[0])
    m,_ := strconv.Atoi(ymd[1])
    d,_ := strconv.Atoi(ymd[2])
    return y,m,d
}

func appendODL(OccupiedDaysList map[int][]OccupiedDay, month int, startDay int, endDay int, colorID string) {
    for day := startDay; day < endDay; day ++{
        OccupiedDaysList[month]= append(OccupiedDaysList[month], OccupiedDay{day, colorIdDict[colorID]})
    }
}

func refreshOccupiedDaysList() (map[int][]OccupiedDay, []string) {
    OccupiedDaysList := make(map[int][]OccupiedDay)
    var EventList []string
    t := time.Now().Format(time.RFC3339)
    events, err := calendarService.Events.List("primary").ShowDeleted(false).
        SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
    if err != nil {
        log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
    }

    fmt.Println("Upcoming events:")
    if len(events.Items) == 0 {
            fmt.Println("No upcoming events found.")
    } else {
        for _, item := range events.Items {
            colorID := item.ColorId
            
            startDate := item.Start.Date
            if startDate == "" {
                fmt.Printf("**NOT ADDED : on %v\n\tOnly all-day events will appear in calendar\n", item.Start.DateTime)
            } else{
                fmt.Printf("# %v (%v)\n", item.Summary, startDate)

                y, m, d := getYMD(startDate)            
                OccupiedDaysList[m]= append(OccupiedDaysList[m], OccupiedDay{d, colorIdDict[colorID]})
                fmt.Printf("\tSTART DATE: year: %v, month: %v, day: %v \n", y, m, d)

                endDate :=item.End.Date
                if endDate != startDate {
                    endY, endM,endD := getYMD(endDate)
                    fmt.Printf("\tENDDATE: year: %v, month: %v, day: %v \n", endY, endM, endD)
                    
                    if m == endM{
                        appendODL(OccupiedDaysList, m, d, endD, colorID)
                        
                    } else{ 
                        appendODL(OccupiedDaysList,m, d, 32, colorID)
                        for month := m+1; month < endM; month ++{
                            appendODL(OccupiedDaysList, month, 0, 32, colorID)
                        }
                        appendODL(OccupiedDaysList, endM, 0, endD, colorID)
                    }    
                }

                fmt.Printf("\tcolorID: '%v' aka %v \n",colorID, colorIdDict[colorID])

                EventList = append(EventList, "Du "+startDate+" au "+endDate+ " : "+ item.Summary)
                // end foreach items
            }
        }
    }

    fmt.Println("-------------\nClick Enter to exit\n")
    return OccupiedDaysList, EventList
}



// start and stop Http server
func startHttpServer(wg *sync.WaitGroup) *http.Server{
    log.Printf("main: starting HTTP server")

    wg.Add(1)
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
        fmt.Println("PORT environment variable was not defined, so the 8080 port is used instead.")
    }

    httpServer := &http.Server{Addr: ":"+port}

    http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources")))) 
    http.HandleFunc("/", handler)

    go func() {
        defer wg.Done() // let main know we are done cleaning up

        // always returns error. ErrServerClosed on graceful close
        if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
            // unexpected error. port in use?
            log.Fatalf("ListenAndServe(): %v", err)
        }
    }()
    return(httpServer)

}



func stopHttpServer(wg *sync.WaitGroup, httpServer *http.Server){
    log.Printf("main: stopping HTTP server")

    // now close the server gracefully ("shutdown")
    
    if err := httpServer.Shutdown(ctx); err != nil {
        panic(err) // failure/timeout shutting down the server gracefully
    }

    // wait for goroutine started in startHttpServer() to stop
    wg.Wait()

    log.Printf("main: done. exiting")
}

var oauth2Config *oauth2.Config;


func main() {
    ctx = context.Background()
    //read calendar
    // file, _ := os.Create("./temp.txt")
    // writer := bufio.NewWriter(file)
    // writer.WriteString("STARTING\n" )

    calendarServiceStarted =false
    tokenRequested =false
 
    wg := &sync.WaitGroup{}
    // start http server to display calendar
    //writer.WriteString("start server\n" )
    httpServer := startHttpServer(wg)
    //writer.WriteString("refresh\n" )
    
    //writer.WriteString("read line \n" )

    // TO retrieve token on first connect you need to comment the following lines
    // keeping only the time.Sleep
    reader := bufio.NewReader(os.Stdin)
    _, err := reader.ReadString('\n')

    if err == io.EOF {
        log.Printf("EOF error have occurred\n")
        time.Sleep(2*time.Minute)
        //writer.WriteString("EOF \n" )
        
    } else if err != nil {
        //writer.WriteString("fatal error \n" )
        log.Fatal(err)
    } 
    
    stopHttpServer(wg,httpServer)
    // writer.WriteString("END \n" )
    // writer.Flush()
}
