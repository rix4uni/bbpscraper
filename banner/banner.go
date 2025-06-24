package banner

import (
	"fmt"
)

// prints the version message
const version = "v0.0.1"

func PrintVersion() {
	fmt.Printf("Current bbpscraper version %s\n", version)
}

// Prints the Colorful banner
func PrintBanner() {
	banner := `
    __     __                                                      
   / /_   / /_   ____   _____ _____ _____ ____ _ ____   ___   _____
  / __ \ / __ \ / __ \ / ___// ___// ___// __  // __ \ / _ \ / ___/
 / /_/ // /_/ // /_/ /(__  )/ /__ / /   / /_/ // /_/ //  __// /    
/_.___//_.___// .___//____/ \___//_/    \__,_// .___/ \___//_/     
             /_/                             /_/`
	fmt.Printf("%s\n%65s\n\n", banner, "Current bbpscraper version "+version)
}
