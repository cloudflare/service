// Package log implements leveled logging. Although the API is inspired by
// https://github.com/golang/glog which aims to be analogous to the Google-internal
// C++ INFO/ERROR/V setup the behavior is different.
//
// It provides functions Info, Warning, Error, Fatal, plus formatting variants such as
// Infof. These functions are protected by the verbosity set via the -v flag.
// All logs are sent to standard error.
//
// Basic examples:
//
//	log.Info("Prepare to repel boarders")
//
//	log.Fatalf("Initialization failed: %s", err)
//
// All log statements are written to standard error.
// This package uses flags for configuration. As a result, flag.Parse must be called.
//
//	-log_backtrace_at=""
//		When set to a file and line number holding a logging statement,
//		such as
//			-log_backtrace_at=gopherflakes.go:234
//		a stack trace will be written to the Info log whenever execution
//		hits that statement. (Unlike with -vmodule, the ".go" must be
//		present.)
//	-v="info"
//		Enable logging at the specified level and above.
//
package log
