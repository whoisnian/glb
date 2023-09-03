package ansi

// color
const (
	// foreground
	BlackFG   string = "\x1b[30m"
	RedFG     string = "\x1b[31m"
	GreenFG   string = "\x1b[32m"
	YellowFG  string = "\x1b[33m"
	BlueFG    string = "\x1b[34m"
	MagentaFG string = "\x1b[35m"
	CyanFG    string = "\x1b[36m"
	WhiteFG   string = "\x1b[37m"

	// background
	BlackBG   string = "\x1b[40m"
	RedBG     string = "\x1b[41m"
	GreenBG   string = "\x1b[42m"
	YellowBG  string = "\x1b[43m"
	BlueBG    string = "\x1b[44m"
	MagentaBG string = "\x1b[45m"
	CyanBG    string = "\x1b[46m"
	WhiteBG   string = "\x1b[47m"

	// other
	Reset        string = "\x1b[0m"
	Bold         string = "\x1b[1m"
	Faint        string = "\x1b[2m"
	Italic       string = "\x1b[3m"
	Underline    string = "\x1b[4m"
	BlinkSlow    string = "\x1b[5m"
	BlinkRapid   string = "\x1b[6m"
	ReverseVideo string = "\x1b[7m"
	Concealed    string = "\x1b[8m"
	CrossedOut   string = "\x1b[9m"
)
