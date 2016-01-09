/*
Package rgo provides a mechanism to call into R. This package assumes that
you have the R binary in your PATH and that you have the jsonlite and RCurl
R packages installed.

Why make rgo?

R has many useful plotting and statistics libraries. rgo is a simple libary
that gives you access to these from Go. This way, you can transform and
process your data in Go, and use R's plotting/statistics libraries to
generate your final figures. As a result, this package is not designed to
be high-performance or have minimal memory consumption. Transferring data
between Go and R requires making several copies of the data, and memory
will not be freed in R. Nevertheless, this package should be sufficient for
basic R usage.
*/
package rgo
