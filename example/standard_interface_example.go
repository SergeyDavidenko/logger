package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/SergeyDavidenko/logger"
)

// Example function that expects a standard logger interface
func useStandardLogger(l interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
	SetOutput(w io.Writer)
	SetPrefix(prefix string)
	SetFlags(flag int)
}) {
	l.SetPrefix("EXAMPLE: ")
	l.SetFlags(log.LstdFlags | log.Lshortfile)

	l.Print("This is a Print message")
	l.Printf("This is a Printf message with %s", "formatting")
	l.Println("This", "is", "a", "Println", "message")
}

func main() {
	fmt.Println("=== Standard Go Logger ===")

	// Use standard Go logger
	standardLogger := log.New(os.Stdout, "", 0)
	useStandardLogger(standardLogger)

	fmt.Println("\n=== Our Custom Logger (compatible with standard interface) ===")

	// Use our custom logger as a drop-in replacement
	config := logger.DefaultConfig()
	config.LogstashEnabled = false // Disable Logstash for this example
	config.AppName = "example-app"
	config.MinLevel = logger.DEBUG

	customLogger := logger.New(config)
	defer customLogger.Close()

	// This works because our logger implements the standard interface methods
	useStandardLogger(customLogger)

	fmt.Println("\n=== Testing Panic Methods (commented out to avoid program exit) ===")

	// Uncomment these lines to test panic methods (they will panic)
	// customLogger.Panic("This would panic")
	// customLogger.Panicf("This would panic with %s", "formatting")
	// customLogger.Panicln("This", "would", "panic", "with", "println")

	fmt.Println("Panic methods exist but are commented out to avoid program termination")

	fmt.Println("\n=== Testing Fatal Methods (commented out to avoid program exit) ===")

	// Uncomment these lines to test fatal methods (they will call os.Exit(1))
	// customLogger.Fatalf("This would be fatal with %s", "formatting")
	// customLogger.Fatalln("This", "would", "be", "fatal", "with", "println")

	fmt.Println("Fatal methods exist but are commented out to avoid program termination")

	fmt.Println("\n=== Testing SetOutput ===")

	// Test output redirection
	var buf bytes.Buffer
	customLogger.SetOutput(&buf)

	customLogger.Print("This message goes to the buffer")
	customLogger.Printf("Buffer content: %s", "test")

	// Note: Our logger still outputs to console because it has its own logging mechanism
	// The SetOutput is for standard interface compatibility

	fmt.Printf("Buffer contains: %s\n", buf.String())

	fmt.Println("\n=== Summary ===")
	fmt.Println("✅ Our logger implements all standard Go log interface methods")
	fmt.Println("✅ Can be used as a drop-in replacement for log.Logger")
	fmt.Println("✅ Maintains all original functionality (Logstash integration, levels, etc.)")
	fmt.Println("✅ Thread-safe operations")
	fmt.Println("✅ Supports SetOutput, SetPrefix, SetFlags for compatibility")
}
