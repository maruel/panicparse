// Copyright 2017 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// panic crashes in various ways.
//
// It is a tool to help test pp, it is used in its unit tests.
//
// To install, run:
//   go install github.com/maruel/panicparse/v2/cmd/panic
//   panic -help
//   panic str |& pp
//
// Some panics require the race detector with -race:
//   go install -race github.com/maruel/panicparse/v2/cmd/panic
//   panic race |& pp
//
// To use with optimization (-N) and inlining (-l) disabled, build with
// -gcflags '-N -l' like:
//   go install -gcflags '-N -l' github.com/maruel/panicparse/v2/cmd/panic
package main

// To add a new panic stack signature, add it to types type below, keeping the
// list ordered by name. If you need utility functions, add it in the section
// below. That's it!

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/maruel/panicparse/v2/cmd/panic/internal"
	correct "github.com/maruel/panicparse/v2/cmd/panic/internal/incorrect"
	ùtf8 "github.com/maruel/panicparse/v2/cmd/panic/internal/utf8"
)

func main() {
	if len(os.Args) == 2 {
		switch n := os.Args[1]; n {
		case "-h", "-help", "--help", "help":
			usage()
			os.Exit(0)

		case "dump_commands":
			// Undocumented command to do a raw dump of the supported commands. This
			// is used by unit tests in ../../stack.
			items := make([]string, 0, len(types))
			for n := range types {
				items = append(items, n)
			}
			sort.Strings(items)
			for _, n := range items {
				fmt.Printf("%s\n", n)
			}
			os.Exit(0)

		default:
			if f, ok := types[n]; ok {
				fmt.Printf("GOTRACEBACK=%s\n", os.Getenv("GOTRACEBACK"))
				if n == "simple" {
					// Since the map lookup creates another call stack entry, add a
					// one-off "simple" panic style to test the very minimal case.
					// types["simple"].f is never called.
					panic("simple")
				}
				f.f()
				os.Exit(3)
			}
			fmt.Fprintf(stdErr, "unknown panic style %q\n", n)
			os.Exit(1)
		}
	}
	usage()
	os.Exit(1)
}

// Mocked in test.
var stdErr io.Writer = os.Stderr

// Utility functions.

func panicint(i int) {
	panic(i)
}

func panicfloat64(f float64) {
	panic(f)
}

func panicstr(a string) {
	panic(a)
}

func panicslicestr(a []string) {
	panic(a)
}

func panicArgsElided(a, b, c, d, e, f, g, h, i, j, k int) {
	panic(a)
}

func recurse(i int) {
	if i > 0 {
		recurse(i - 1)
		return
	}
	panic(42)
}

func panicRaceDisabled(name string) {
	help := "'panic %s' can only be used when built with the race detector.\n" +
		"To build, use:\n" +
		"  go install -race github.com/maruel/panicparse/v2/cmd/panic\n"
	fmt.Fprintf(stdErr, help, name)
}

func rerunWithFastCrash() {
	if os.Getenv("GORACE") != "log_path=stderr halt_on_error=1" {
		_ = os.Setenv("GORACE", "log_path=stderr halt_on_error=1")
		c := exec.Command(os.Args[0], os.Args[1:]...)
		c.Stderr = os.Stderr
		if err, ok := c.Run().(*exec.ExitError); ok {
			if status, ok := err.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
			os.Exit(1)
		}
		os.Exit(0)
	}
}

// panicDoRaceWrite and panicDoRaceRead are extracted from panicRace() to make
// the stack trace less trivial, but in general folks will do the error with
// this code inlined.
func panicDoRaceWrite(x *int) {
	for i := 0; ; i++ {
		*x = i
	}
}
func panicDoRaceRead(x *int) {
	for i := 0; ; {
		i += *x
	}
}

func panicRace() {
	if !raceEnabled {
		panicRaceDisabled("race")
		return
	}
	rerunWithFastCrash()

	i := 0
	// Do two separate calls so that the 'created at' stacks are different.
	go func() {
		panicDoRaceWrite(&i)
	}()
	go func() {
		panicDoRaceRead(&i)
	}()
	time.Sleep(time.Minute)
}

//go:noinline
func panicChanStruct(x chan struct{}) {
	panic("test")
}

/* TODO(maruel): This is not detected!
func panicRaceUnaligned() {
	if !raceEnabled {
		panicRaceDisabled("race_unaligned")
		return
	}
	rerunWithFastCrash()

	a := [8]byte{}
	b := (*int64)(unsafe.Pointer(&a[0]))
	go func() {
		for i := 0; ; i++ {
			a[4] = byte(i)
		}
	}()
	go func() {
		for {
			*b++
		}
	}()
	time.Sleep(time.Minute)
}
*/

//

// types is all the supported types of panics.
//
// Keep the list sorted.
//
// TODO(maruel): Figure out a way to reliably trigger "(scan)" output:
// - disable automatic GC with runtime.SetGCPercent(-1)
// - a goroutine with a large number of items in the stack
// - large heap to make the scanning process slow enough
// - trigger a manual GC with go runtime.GC()
// - panic in the meantime
// This would still not be deterministic.
//
// TODO(maruel): Figure out a way to reliably trigger sleep output.
var types = map[string]struct {
	desc string
	f    func()
}{
	"args_elided": {
		"too many args in stack line, causing the call arguments to be elided",
		func() {
			panicArgsElided(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
		},
	},

	"chan_receive": {
		"goroutine blocked on <-c",
		func() {
			c := make(chan bool)
			go func() {
				<-c
				<-c
			}()
			c <- true
			panic(42)
		},
	},

	"chan_send": {
		"goroutine blocked on c<-",
		func() {
			c := make(chan bool)
			go func() {
				c <- true
				c <- true
			}()
			<-c
			panic(42)
		},
	},

	"chan_struct": {
		"panic with an empty chan struct{} as a parameter",
		func() {
			panicChanStruct(nil)
		},
	},

	"float": {
		"panic(4.2)",
		func() {
			panicfloat64(4.2)
		},
	},

	"goroutine_1": {
		"panic in one goroutine",
		func() {
			go func() {
				panicint(42)
			}()
			time.Sleep(time.Minute)
		},
	},

	"goroutine_100": {
		"start 100 goroutines before panicking",
		func() {
			var wg sync.WaitGroup
			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func() {
					wg.Done()
					time.Sleep(time.Minute)
				}()
			}
			wg.Wait()
			panicint(42)
		},
	},

	"goroutine_dedupe_pointers": {
		"start 100 goroutines with different pointers before panicking",
		func() {
			var wg sync.WaitGroup
			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func(b *int) {
					wg.Done()
					time.Sleep(time.Minute)
				}(new(int))
			}
			wg.Wait()
			panicint(42)
		},
	},

	"int": {
		"panic(42)",
		func() {
			panicint(42)
		},
	},

	"locked": {
		"thread locked goroutine via runtime.LockOSThread()",
		func() {
			runtime.LockOSThread()
			panic(42)
		},
	},

	"other": {
		"panics with other package in the call stack, with both exported and unexpected functions",
		func() {
			internal.Callback(func() {
				panic("allo")
			})
		},
	},

	"asleep": {
		"panics with 'all goroutines are asleep - deadlock'",
		func() {
			// When built with the race detector, this hangs. I suspect this is due
			// because the race detector starts a separate goroutine which causes
			// checkdead() to not trigger. See checkdead() in src/runtime/proc.go.
			// https://github.com/golang/go/issues/20588
			//
			// Repro:
			//   go install -race github.com/maruel/panicparse/v2/cmd/panic; panic asleep
			var mu sync.Mutex
			mu.Lock()
			mu.Lock()
		},
	},

	"race": {
		"cause a crash by race detector",
		panicRace,
	},

	/* TODO(maruel): This is not detected!
	"race_unaligned": {
		"cause a crash by race detector with unaligned access",
		panicRaceUnaligned,
	},
	*/

	"stack_cut_off": {
		"recursive calls with too many call lines in traceback, causing higher up calls to missing",
		func() {
			// Observed limit is 99.
			// src/runtime/runtime2.go:const _TracebackMaxFrames = 100
			recurse(100)
		},
	},

	"stack_cut_off_named": {
		"named calls with too many call lines in traceback, causing higher up calls to missing",
		func() {
			// The tool has difficulty with very deep static calls during linking.
			// See https://github.com/golang/go/issues/51814
			recurse497()
		},
	},

	"simple": {
		// This is not used for real, here for documentation.
		"skip the map for a shorter stack trace",
		func() {},
	},

	"slice_str": {
		"panic([]string{\"allo\"}) with cap=2",
		func() {
			a := make([]string, 1, 2)
			a[0] = "allo"
			panicslicestr(a)
		},
	},

	"stdlib": {
		"panics with stdlib in the call stack, with both exported and unexpected functions",
		func() {
			strings.FieldsFunc("a", func(rune) bool {
				panic("allo")
			})
		},
	},

	"stdlib_and_other": {
		"panics with both other and stdlib packages in the call stack",
		func() {
			strings.FieldsFunc("a", func(rune) bool {
				internal.Callback(func() {
					panic("allo")
				})
				return false
			})
		},
	},

	"str": {
		"panic(\"allo\")",
		func() {
			panicstr("allo")
		},
	},

	"mismatched": {
		"mismatched package and directory names",
		func() {
			correct.Panic()
		},
	},

	"utf8": {
		"non-ascii package, struct and method names",
		func() {
			s := ùtf8.Strùct{}
			s.Pànic()
		},
	},
}

func usage() {
	t := `usage: panic <way>

This tool is meant to be used with pp to test different parsing scenarios and
ensure output on different version of the Go toolchain can be successfully
parsed.

Set GOTRACEBACK before running this tool to see how it affects the panic output.

Built with: ` + runtime.Version() + `

Select the way to panic:
`
	_, _ = io.WriteString(stdErr, t)
	names := make([]string, 0, len(types))
	m := 0
	for n := range types {
		names = append(names, n)
		if i := len(n); i > m {
			m = i
		}
	}
	sort.Strings(names)
	for _, n := range names {
		fmt.Fprintf(stdErr, "- %-*s  %s\n", m, n, types[n].desc)
	}
}

//

func recurse00()  { panic("the end") }
func recurse01()  { recurse00() }
func recurse02()  { recurse01() }
func recurse03()  { recurse02() }
func recurse04()  { recurse03() }
func recurse05()  { recurse04() }
func recurse06()  { recurse05() }
func recurse07()  { recurse06() }
func recurse08()  { recurse07() }
func recurse09()  { recurse08() }
func recurse10()  { recurse09() }
func recurse11()  { recurse10() }
func recurse12()  { recurse11() }
func recurse13()  { recurse12() }
func recurse14()  { recurse13() }
func recurse15()  { recurse14() }
func recurse16()  { recurse15() }
func recurse17()  { recurse16() }
func recurse18()  { recurse17() }
func recurse19()  { recurse18() }
func recurse20()  { recurse19() }
func recurse21()  { recurse20() }
func recurse22()  { recurse21() }
func recurse23()  { recurse22() }
func recurse24()  { recurse23() }
func recurse25()  { recurse24() }
func recurse26()  { recurse25() }
func recurse27()  { recurse26() }
func recurse28()  { recurse27() }
func recurse29()  { recurse28() }
func recurse30()  { recurse29() }
func recurse31()  { recurse30() }
func recurse32()  { recurse31() }
func recurse33()  { recurse32() }
func recurse34()  { recurse33() }
func recurse35()  { recurse34() }
func recurse36()  { recurse35() }
func recurse37()  { recurse36() }
func recurse38()  { recurse37() }
func recurse39()  { recurse38() }
func recurse40()  { recurse39() }
func recurse41()  { recurse40() }
func recurse42()  { recurse41() }
func recurse43()  { recurse42() }
func recurse44()  { recurse43() }
func recurse45()  { recurse44() }
func recurse46()  { recurse45() }
func recurse47()  { recurse46() }
func recurse48()  { recurse47() }
func recurse49()  { recurse48() }
func recurse50()  { recurse49() }
func recurse51()  { recurse50() }
func recurse52()  { recurse51() }
func recurse53()  { recurse52() }
func recurse54()  { recurse53() }
func recurse55()  { recurse54() }
func recurse56()  { recurse55() }
func recurse57()  { recurse56() }
func recurse58()  { recurse57() }
func recurse59()  { recurse58() }
func recurse60()  { recurse59() }
func recurse61()  { recurse60() }
func recurse62()  { recurse61() }
func recurse63()  { recurse62() }
func recurse64()  { recurse63() }
func recurse65()  { recurse64() }
func recurse66()  { recurse65() }
func recurse67()  { recurse66() }
func recurse68()  { recurse67() }
func recurse69()  { recurse68() }
func recurse70()  { recurse69() }
func recurse71()  { recurse70() }
func recurse72()  { recurse71() }
func recurse73()  { recurse72() }
func recurse74()  { recurse73() }
func recurse75()  { recurse74() }
func recurse76()  { recurse75() }
func recurse77()  { recurse76() }
func recurse78()  { recurse77() }
func recurse79()  { recurse78() }
func recurse80()  { recurse79() }
func recurse81()  { recurse80() }
func recurse82()  { recurse81() }
func recurse83()  { recurse82() }
func recurse84()  { recurse83() }
func recurse85()  { recurse84() }
func recurse86()  { recurse85() }
func recurse87()  { recurse86() }
func recurse88()  { recurse87() }
func recurse89()  { recurse88() }
func recurse90()  { recurse89() }
func recurse91()  { recurse90() }
func recurse92()  { recurse91() }
func recurse93()  { recurse92() }
func recurse94()  { recurse93() }
func recurse95()  { recurse94() }
func recurse96()  { recurse95() }
func recurse97()  { recurse96() }
func recurse98()  { recurse97() }
func recurse99()  { recurse98() }
func recurse100() { recurse99() }

func recurse101() { recurse100() }
func recurse102() { recurse101() }
func recurse103() { recurse102() }
func recurse104() { recurse103() }
func recurse105() { recurse104() }
func recurse106() { recurse105() }
func recurse107() { recurse106() }
func recurse108() { recurse107() }
func recurse109() { recurse108() }
func recurse110() { recurse109() }
func recurse111() { recurse110() }
func recurse112() { recurse111() }
func recurse113() { recurse112() }
func recurse114() { recurse113() }
func recurse115() { recurse114() }
func recurse116() { recurse115() }
func recurse117() { recurse116() }
func recurse118() { recurse117() }
func recurse119() { recurse118() }
func recurse120() { recurse119() }
func recurse121() { recurse120() }
func recurse122() { recurse121() }
func recurse123() { recurse122() }
func recurse124() { recurse123() }
func recurse125() { recurse124() }
func recurse126() { recurse125() }
func recurse127() { recurse126() }
func recurse128() { recurse127() }
func recurse129() { recurse128() }
func recurse130() { recurse129() }
func recurse131() { recurse130() }
func recurse132() { recurse131() }
func recurse133() { recurse132() }
func recurse134() { recurse133() }
func recurse135() { recurse134() }
func recurse136() { recurse135() }
func recurse137() { recurse136() }
func recurse138() { recurse137() }
func recurse139() { recurse138() }
func recurse140() { recurse139() }
func recurse141() { recurse140() }
func recurse142() { recurse141() }
func recurse143() { recurse142() }
func recurse144() { recurse143() }
func recurse145() { recurse144() }
func recurse146() { recurse145() }
func recurse147() { recurse146() }
func recurse148() { recurse147() }
func recurse149() { recurse148() }
func recurse150() { recurse149() }
func recurse151() { recurse150() }
func recurse152() { recurse151() }
func recurse153() { recurse152() }
func recurse154() { recurse153() }
func recurse155() { recurse154() }
func recurse156() { recurse155() }
func recurse157() { recurse156() }
func recurse158() { recurse157() }
func recurse159() { recurse158() }
func recurse160() { recurse159() }
func recurse161() { recurse160() }
func recurse162() { recurse161() }
func recurse163() { recurse162() }
func recurse164() { recurse163() }
func recurse165() { recurse164() }
func recurse166() { recurse165() }
func recurse167() { recurse166() }
func recurse168() { recurse167() }
func recurse169() { recurse168() }
func recurse170() { recurse169() }
func recurse171() { recurse170() }
func recurse172() { recurse171() }
func recurse173() { recurse172() }
func recurse174() { recurse173() }
func recurse175() { recurse174() }
func recurse176() { recurse175() }
func recurse177() { recurse176() }
func recurse178() { recurse177() }
func recurse179() { recurse178() }
func recurse180() { recurse179() }
func recurse181() { recurse180() }
func recurse182() { recurse181() }
func recurse183() { recurse182() }
func recurse184() { recurse183() }
func recurse185() { recurse184() }
func recurse186() { recurse185() }
func recurse187() { recurse186() }
func recurse188() { recurse187() }
func recurse189() { recurse188() }
func recurse190() { recurse189() }
func recurse191() { recurse190() }
func recurse192() { recurse191() }
func recurse193() { recurse192() }
func recurse194() { recurse193() }
func recurse195() { recurse194() }
func recurse196() { recurse195() }
func recurse197() { recurse196() }
func recurse198() { recurse197() }
func recurse199() { recurse198() }

func recurse200() { recurse199() }
func recurse201() { recurse200() }
func recurse202() { recurse201() }
func recurse203() { recurse202() }
func recurse204() { recurse203() }
func recurse205() { recurse204() }
func recurse206() { recurse205() }
func recurse207() { recurse206() }
func recurse208() { recurse207() }
func recurse209() { recurse208() }
func recurse210() { recurse209() }
func recurse211() { recurse210() }
func recurse212() { recurse211() }
func recurse213() { recurse212() }
func recurse214() { recurse213() }
func recurse215() { recurse214() }
func recurse216() { recurse215() }
func recurse217() { recurse216() }
func recurse218() { recurse217() }
func recurse219() { recurse218() }
func recurse220() { recurse219() }
func recurse221() { recurse220() }
func recurse222() { recurse221() }
func recurse223() { recurse222() }
func recurse224() { recurse223() }
func recurse225() { recurse224() }
func recurse226() { recurse225() }
func recurse227() { recurse226() }
func recurse228() { recurse227() }
func recurse229() { recurse228() }
func recurse230() { recurse229() }
func recurse231() { recurse230() }
func recurse232() { recurse231() }
func recurse233() { recurse232() }
func recurse234() { recurse233() }
func recurse235() { recurse234() }
func recurse236() { recurse235() }
func recurse237() { recurse236() }
func recurse238() { recurse237() }
func recurse239() { recurse238() }
func recurse240() { recurse239() }
func recurse241() { recurse240() }
func recurse242() { recurse241() }
func recurse243() { recurse242() }
func recurse244() { recurse243() }
func recurse245() { recurse244() }
func recurse246() { recurse245() }
func recurse247() { recurse246() }
func recurse248() { recurse247() }
func recurse249() { recurse248() }
func recurse250() { recurse249() }
func recurse251() { recurse250() }
func recurse252() { recurse251() }
func recurse253() { recurse252() }
func recurse254() { recurse253() }
func recurse255() { recurse254() }
func recurse256() { recurse255() }
func recurse257() { recurse256() }
func recurse258() { recurse257() }
func recurse259() { recurse258() }
func recurse260() { recurse259() }
func recurse261() { recurse260() }
func recurse262() { recurse261() }
func recurse263() { recurse262() }
func recurse264() { recurse263() }
func recurse265() { recurse264() }
func recurse266() { recurse265() }
func recurse267() { recurse266() }
func recurse268() { recurse267() }
func recurse269() { recurse268() }
func recurse270() { recurse269() }
func recurse271() { recurse270() }
func recurse272() { recurse271() }
func recurse273() { recurse272() }
func recurse274() { recurse273() }
func recurse275() { recurse274() }
func recurse276() { recurse275() }
func recurse277() { recurse276() }
func recurse278() { recurse277() }
func recurse279() { recurse278() }
func recurse280() { recurse279() }
func recurse281() { recurse280() }
func recurse282() { recurse281() }
func recurse283() { recurse282() }
func recurse284() { recurse283() }
func recurse285() { recurse284() }
func recurse286() { recurse285() }
func recurse287() { recurse286() }
func recurse288() { recurse287() }
func recurse289() { recurse288() }
func recurse290() { recurse289() }
func recurse291() { recurse290() }
func recurse292() { recurse291() }
func recurse293() { recurse292() }
func recurse294() { recurse293() }
func recurse295() { recurse294() }
func recurse296() { recurse295() }
func recurse297() { recurse296() }
func recurse298() { recurse297() }
func recurse299() { recurse298() }

func recurse300() { recurse299() }
func recurse301() { recurse300() }
func recurse302() { recurse301() }
func recurse303() { recurse302() }
func recurse304() { recurse303() }
func recurse305() { recurse304() }
func recurse306() { recurse305() }
func recurse307() { recurse306() }
func recurse308() { recurse307() }
func recurse309() { recurse308() }
func recurse310() { recurse309() }
func recurse311() { recurse310() }
func recurse312() { recurse311() }
func recurse313() { recurse312() }
func recurse314() { recurse313() }
func recurse315() { recurse314() }
func recurse316() { recurse315() }
func recurse317() { recurse316() }
func recurse318() { recurse317() }
func recurse319() { recurse318() }
func recurse320() { recurse319() }
func recurse321() { recurse320() }
func recurse322() { recurse321() }
func recurse323() { recurse322() }
func recurse324() { recurse323() }
func recurse325() { recurse324() }
func recurse326() { recurse325() }
func recurse327() { recurse326() }
func recurse328() { recurse327() }
func recurse329() { recurse328() }
func recurse330() { recurse329() }
func recurse331() { recurse330() }
func recurse332() { recurse331() }
func recurse333() { recurse332() }
func recurse334() { recurse333() }
func recurse335() { recurse334() }
func recurse336() { recurse335() }
func recurse337() { recurse336() }
func recurse338() { recurse337() }
func recurse339() { recurse338() }
func recurse340() { recurse339() }
func recurse341() { recurse340() }
func recurse342() { recurse341() }
func recurse343() { recurse342() }
func recurse344() { recurse343() }
func recurse345() { recurse344() }
func recurse346() { recurse345() }
func recurse347() { recurse346() }
func recurse348() { recurse347() }
func recurse349() { recurse348() }
func recurse350() { recurse349() }
func recurse351() { recurse350() }
func recurse352() { recurse351() }
func recurse353() { recurse352() }
func recurse354() { recurse353() }
func recurse355() { recurse354() }
func recurse356() { recurse355() }
func recurse357() { recurse356() }
func recurse358() { recurse357() }
func recurse359() { recurse358() }
func recurse360() { recurse359() }
func recurse361() { recurse360() }
func recurse362() { recurse361() }
func recurse363() { recurse362() }
func recurse364() { recurse363() }
func recurse365() { recurse364() }
func recurse366() { recurse365() }
func recurse367() { recurse366() }
func recurse368() { recurse367() }
func recurse369() { recurse368() }
func recurse370() { recurse369() }
func recurse371() { recurse370() }
func recurse372() { recurse371() }
func recurse373() { recurse372() }
func recurse374() { recurse373() }
func recurse375() { recurse374() }
func recurse376() { recurse375() }
func recurse377() { recurse376() }
func recurse378() { recurse377() }
func recurse379() { recurse378() }
func recurse380() { recurse379() }
func recurse381() { recurse380() }
func recurse382() { recurse381() }
func recurse383() { recurse382() }
func recurse384() { recurse383() }
func recurse385() { recurse384() }
func recurse386() { recurse385() }
func recurse387() { recurse386() }
func recurse388() { recurse387() }
func recurse389() { recurse388() }
func recurse390() { recurse389() }
func recurse391() { recurse390() }
func recurse392() { recurse391() }
func recurse393() { recurse392() }
func recurse394() { recurse393() }
func recurse395() { recurse394() }
func recurse396() { recurse395() }
func recurse397() { recurse396() }
func recurse398() { recurse397() }
func recurse399() { recurse398() }

func recurse400() { recurse399() }
func recurse401() { recurse400() }
func recurse402() { recurse401() }
func recurse403() { recurse402() }
func recurse404() { recurse403() }
func recurse405() { recurse404() }
func recurse406() { recurse405() }
func recurse407() { recurse406() }
func recurse408() { recurse407() }
func recurse409() { recurse408() }
func recurse410() { recurse409() }
func recurse411() { recurse410() }
func recurse412() { recurse411() }
func recurse413() { recurse412() }
func recurse414() { recurse413() }
func recurse415() { recurse414() }
func recurse416() { recurse415() }
func recurse417() { recurse416() }
func recurse418() { recurse417() }
func recurse419() { recurse418() }
func recurse420() { recurse419() }
func recurse421() { recurse420() }
func recurse422() { recurse421() }
func recurse423() { recurse422() }
func recurse424() { recurse423() }
func recurse425() { recurse424() }
func recurse426() { recurse425() }
func recurse427() { recurse426() }
func recurse428() { recurse427() }
func recurse429() { recurse428() }
func recurse430() { recurse429() }
func recurse431() { recurse430() }
func recurse432() { recurse431() }
func recurse433() { recurse432() }
func recurse434() { recurse433() }
func recurse435() { recurse434() }
func recurse436() { recurse435() }
func recurse437() { recurse436() }
func recurse438() { recurse437() }
func recurse439() { recurse438() }
func recurse440() { recurse439() }
func recurse441() { recurse440() }
func recurse442() { recurse441() }
func recurse443() { recurse442() }
func recurse444() { recurse443() }
func recurse445() { recurse444() }
func recurse446() { recurse445() }
func recurse447() { recurse446() }
func recurse448() { recurse447() }
func recurse449() { recurse448() }
func recurse450() { recurse449() }
func recurse451() { recurse450() }
func recurse452() { recurse451() }
func recurse453() { recurse452() }
func recurse454() { recurse453() }
func recurse455() { recurse454() }
func recurse456() { recurse455() }
func recurse457() { recurse456() }
func recurse458() { recurse457() }
func recurse459() { recurse458() }
func recurse460() { recurse459() }
func recurse461() { recurse460() }
func recurse462() { recurse461() }
func recurse463() { recurse462() }
func recurse464() { recurse463() }
func recurse465() { recurse464() }
func recurse466() { recurse465() }
func recurse467() { recurse466() }
func recurse468() { recurse467() }
func recurse469() { recurse468() }
func recurse470() { recurse469() }
func recurse471() { recurse470() }
func recurse472() { recurse471() }
func recurse473() { recurse472() }
func recurse474() { recurse473() }
func recurse475() { recurse474() }
func recurse476() { recurse475() }
func recurse477() { recurse476() }
func recurse478() { recurse477() }
func recurse479() { recurse478() }
func recurse480() { recurse479() }
func recurse481() { recurse480() }
func recurse482() { recurse481() }
func recurse483() { recurse482() }
func recurse484() { recurse483() }
func recurse485() { recurse484() }
func recurse486() { recurse485() }
func recurse487() { recurse486() }
func recurse488() { recurse487() }
func recurse489() { recurse488() }
func recurse490() { recurse489() }
func recurse491() { recurse490() }
func recurse492() { recurse491() }
func recurse493() { recurse492() }
func recurse494() { recurse493() }
func recurse495() { recurse494() }
func recurse496() { recurse495() }
func recurse497() { recurse496() }
