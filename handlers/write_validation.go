package handlers

// validationStateCheck follows text about an async validation that has not
// finished and names the read tool that reports validation states. It is
// empty while no such tool is served; the file that defines that tool sets
// it in init.
var validationStateCheck = ""
