package appsource

// A sample Shiny app
var sampleApp = `#
# This is a Shiny web application. You can run the application by clicking
# the 'Run App' button above.
#
# Find out more about building applications with Shiny here:
#
#    http://shiny.rstudio.com/
#

library(shiny)

# Define UI for application that draws a histogram
ui <- fluidPage(

    # Application title
    titlePanel("Old Faithful Geyser Data"),

    # Sidebar with a slider input for number of bins 
    sidebarLayout(
        sidebarPanel(
            
            # Greetings
            uiOutput("greetings"),
            tags$br(),
            
            sliderInput("bins",
                        "Number of bins:",
                        min = 1,
                        max = 50,
                        value = 30)
        ),

        # Show a plot of the generated distribution
        mainPanel(
           plotOutput("distPlot")
        )
    )
)

# Define server logic required to draw a histogram
server <- function(input, output, session) {

    output$distPlot <- renderPlot({
        # generate bins based on input$bins from ui.R
        x    <- faithful[, 2]
        bins <- seq(min(x), max(x), length.out = input$bins + 1)

        # draw the histogram with the specified number of bins
        hist(x, breaks = bins, col = 'darkgray', border = 'white')
    })

    output$greetings <- renderUI({
        if (exists("HTTP_APPSERVR_DISPLAYEDNAME", envir = session$request)) {
            tagList(
                tags$h3(paste0("Welcome, ", 
                           get("HTTP_APPSERVR_DISPLAYEDNAME", envir=session$request),
                           "!"), style="margin-top:0;"),
                tags$a("Logout", href="/auth/logout", class="btn btn-primary")
            )
        } else {
            tagList(
                tags$h3("Welcome!", style="margin-top:0;"),
                tags$a("Login", href="/auth/login", class="btn btn-primary")
            )
        }
    })
}

# Run the application 
shinyApp(ui = ui, server = server)`
