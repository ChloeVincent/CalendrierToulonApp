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

    "google.golang.org/api/calendar/v3"
)

type Month struct {  
    Name string
    Index string
    Days  int
    StartDay int
}

type OccupiedPeriod struct {
    Month int
    StartDay int
    EndDay int
    ColorId string
    IsJardin bool
}

type OccupiedDay struct {  
    DayNumber int
    Color string
    IsJardin bool
}

type TemplateArgs struct {
    Months []Month
    OccupiedEvents []string
    JardinEvents []string
}

// initializer function to avoid changes on what should be a constant array
// (if it existed in go)
func getAllMonths() []Month {
    return []Month {
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
}

// initializer function to avoid changes on what should be a constant array
// (if it existed in go)
func getColorIdDict() map[string]string {
    return map[string]string{
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
        if a.DayNumber == e {
            return true
        }
    }
    return false
}

func getOccupiedBorderColor(s []OccupiedDay, e int) string {
    for _, a := range s {
        if a.DayNumber == e && !a.IsJardin {
            return a.Color
        }
    }
    return ""
}

func getClass(month []OccupiedDay, day int, isToday bool) string{
    class:= ""
    for _, d := range month {
        if d.DayNumber == day {
            if d.IsJardin{
                class = class + " jardin"
            } else {
                if isToday{
                    class = class + " today"
                } else{
                    class = class + " occupied"
                }
            }
        }
    }
    if isToday{
        class = class + " today"
    }
    return class
}


// Runs server
func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Println("handler")
    token := getLocalToken(w,r)
    if token == nil{
        fmt.Println("No token was found, redirecting to login")
        http.Redirect(w, r, "/login/", http.StatusSeeOther)
        return
    }
    
    calendarService := startCalendarService(w, token)

    if calendarService == nil{
        log.Fatal("Calendar service was not initialized properly.")
    }
    OccupiedDaysList, occupiedEvents, jardinEvents, now := refreshOccupiedDaysList(calendarService)

    fmap:= template.FuncMap{"Iterate": iterate,
                            "IsOccupied": func(day int, monthStr string) bool {
                                        month, _:= strconv.Atoi(monthStr)
                                        return contains(OccupiedDaysList[month], day)},
                            "DayColor": func(day int, monthStr string) string {
                                        month, _:= strconv.Atoi(monthStr)
                                        return getOccupiedBorderColor(OccupiedDaysList[month], day)},
                            "GetClass": func(day int, monthStr string) string {
                                        month, _:= strconv.Atoi(monthStr)
                                        isToday := (time.Month(month) == now.Month()) && (day == now.Day())
                                        return getClass(OccupiedDaysList[month], day, isToday)},
                            }
    
    t, _ := template.New("template.html").Funcs(fmap).ParseFiles("template.html")
    t.Execute(w, TemplateArgs{getAllMonths(), occupiedEvents, jardinEvents})
}


// refreshOccupiedDaysList functions: getYMD (Year Month Day) and appendODL (OccupiedDaysList)
func getYMD(startDate string)(int, int, int){
    ymd := strings.Split(startDate, "-")
    y,_ := strconv.Atoi(ymd[0])
    m,_ := strconv.Atoi(ymd[1])
    d,_ := strconv.Atoi(ymd[2])
    return y,m,d
}

func appendODL(OccupiedDaysList map[int][]OccupiedDay, period OccupiedPeriod) {
    for day := period.StartDay; day < period.EndDay; day ++{
        OccupiedDaysList[period.Month]= append(OccupiedDaysList[period.Month], 
            OccupiedDay{day, getColorIdDict()[period.ColorId], period.IsJardin})
    }
}

func refreshOccupiedDaysList(calendarService  *calendar.Service) (map[int][]OccupiedDay, []string, []string, time.Time) {
    OccupiedDaysList := make(map[int][]OccupiedDay)
    var EventList, JardinList []string
    tMin := time.Now().AddDate(-1, 9, 0).Format(time.RFC3339)
    tMax := time.Now().AddDate(0, 9, 0).Format(time.RFC3339)
    events, err := calendarService.Events.List("primary").ShowDeleted(false).
        SingleEvents(true).TimeMin(tMin).TimeMax(tMax).OrderBy("startTime").Do()
    if err != nil {
        log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
    }

    fmt.Println("Upcoming events:")
    if len(events.Items) == 0 {
            fmt.Println("No upcoming events found.")
    } else {
        colorIdDict := getColorIdDict()
        for _, item := range events.Items {
            colorID := item.ColorId
            
            startDate := item.Start.Date
            if startDate == "" {
                fmt.Printf("**NOT ADDED : on %v\n\tOnly all-day events will appear in calendar\n", item.Start.DateTime)
            } else{
                fmt.Printf("# %v (%v)\n", item.Summary, startDate)

                isJardin:= strings.HasPrefix(item.Summary, "jardin")

                y, m, d := getYMD(startDate)
                OccupiedDaysList[m]= append(OccupiedDaysList[m], OccupiedDay{d, colorIdDict[colorID], isJardin})
                fmt.Printf("\tSTART DATE: year: %v, month: %v, day: %v \n", y, m, d)

                endDate :=item.End.Date
                if endDate != startDate {
                    endY, endM,endD := getYMD(endDate)
                    fmt.Printf("\tENDDATE: year: %v, month: %v, day: %v \n", endY, endM, endD)
                    
                    if m == endM{
                        appendODL(OccupiedDaysList, OccupiedPeriod{m, d, endD, colorID, isJardin})
                        
                    } else{ 
                        appendODL(OccupiedDaysList,OccupiedPeriod{m, d, 32, colorID, isJardin})
                        for month := m+1; month < endM; month ++{
                            appendODL(OccupiedDaysList, OccupiedPeriod{month, 0, 32, colorID, isJardin})
                        }
                        appendODL(OccupiedDaysList, OccupiedPeriod{endM, 0, endD, colorID, isJardin})
                    }    
                }

                fmt.Printf("\tcolorID: '%v' aka %v \n",colorID, colorIdDict[colorID])

                if isJardin{
                    JardinList = append(JardinList, item.Summary + " ("+startDate+" -- "+endDate +")")
                }else{
                    EventList = append(EventList, "Du "+startDate+" au "+endDate+ " : "+ item.Summary)
                }
                // end foreach items
            }
        }
    }

    fmt.Println("-------------\nClick Enter to exit\n")
    return OccupiedDaysList, EventList, JardinList, time.Now()
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
   http.ServeFile(w, r, "favicon.ico")
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
    http.HandleFunc("/favicon.ico", faviconHandler)
    http.HandleFunc("/login/", loginHandler)
    http.HandleFunc("/oauth2CallBack/", oauth2CallBackHandler)


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
    if err := httpServer.Shutdown(context.Background()); err != nil {
        panic(err) // failure/timeout shutting down the server gracefully
    }

    // wait for goroutine started in startHttpServer() to stop
    wg.Wait()

    log.Printf("main: done. exiting")
}

func main() {
    wg := &sync.WaitGroup{}
    // start http server
    httpServer := startHttpServer(wg)

    // To retrieve token from web on first local connect you need to comment the following lines
    // keeping only the time.Sleep ...
    reader := bufio.NewReader(os.Stdin)
    _, err := reader.ReadString('\n')

    if err == io.EOF {
        log.Printf("EOF error have occurred\n")
        time.Sleep(2*time.Minute)        
    } else if err != nil {
        log.Fatal(err)
    } 

      stopHttpServer(wg,httpServer)
}
