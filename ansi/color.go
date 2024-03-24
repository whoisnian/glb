package ansi

// color
const (
	// foreground
	BlackFG   string = "\x1b[30m" // printf "BlackFG: \x1b[30mBlackFG\x1b[0m\n"
	RedFG     string = "\x1b[31m" // printf "RedFG: \x1b[31mRedFG\x1b[0m\n"
	GreenFG   string = "\x1b[32m" // printf "GreenFG: \x1b[32mGreenFG\x1b[0m\n"
	YellowFG  string = "\x1b[33m" // printf "YellowFG: \x1b[33mYellowFG\x1b[0m\n"
	BlueFG    string = "\x1b[34m" // printf "BlueFG: \x1b[34mBlueFG\x1b[0m\n"
	MagentaFG string = "\x1b[35m" // printf "MagentaFG: \x1b[35mMagentaFG\x1b[0m\n"
	CyanFG    string = "\x1b[36m" // printf "CyanFG: \x1b[36mCyanFG\x1b[0m\n"
	WhiteFG   string = "\x1b[37m" // printf "WhiteFG: \x1b[37mWhiteFG\x1b[0m\n"

	// background
	BlackBG   string = "\x1b[40m" // printf "BlackBG: \x1b[40mBlackBG\x1b[0m\n"
	RedBG     string = "\x1b[41m" // printf "RedBG: \x1b[41mRedBG\x1b[0m\n"
	GreenBG   string = "\x1b[42m" // printf "GreenBG: \x1b[42mGreenBG\x1b[0m\n"
	YellowBG  string = "\x1b[43m" // printf "YellowBG: \x1b[43mYellowBG\x1b[0m\n"
	BlueBG    string = "\x1b[44m" // printf "BlueBG: \x1b[44mBlueBG\x1b[0m\n"
	MagentaBG string = "\x1b[45m" // printf "MagentaBG: \x1b[45mMagentaBG\x1b[0m\n"
	CyanBG    string = "\x1b[46m" // printf "CyanBG: \x1b[46mCyanBG\x1b[0m\n"
	WhiteBG   string = "\x1b[47m" // printf "WhiteBG: \x1b[47mWhiteBG\x1b[0m\n"

	// other
	Reset        string = "\x1b[0m" // printf "Reset: \x1b[0mReset\x1b[0m\n"
	Bold         string = "\x1b[1m" // printf "Bold: \x1b[1mBold\x1b[0m\n"
	Faint        string = "\x1b[2m" // printf "Faint: \x1b[2mFaint\x1b[0m\n"
	Italic       string = "\x1b[3m" // printf "Italic: \x1b[3mItalic\x1b[0m\n"
	Underline    string = "\x1b[4m" // printf "Underline: \x1b[4mUnderline\x1b[0m\n"
	BlinkSlow    string = "\x1b[5m" // printf "BlinkSlow: \x1b[5mBlinkSlow\x1b[0m\n"
	BlinkRapid   string = "\x1b[6m" // printf "BlinkRapid: \x1b[6mBlinkRapid\x1b[0m\n"
	ReverseVideo string = "\x1b[7m" // printf "ReverseVideo: \x1b[7mReverseVideo\x1b[0m\n"
	Concealed    string = "\x1b[8m" // printf "Concealed: \x1b[8mConcealed\x1b[0m\n"
	CrossedOut   string = "\x1b[9m" // printf "CrossedOut: \x1b[9mCrossedOut\x1b[0m\n"
)
