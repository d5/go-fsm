package fsm

// Internal script used to invoke a function from the user script.
//
// input
//   fn  : function name
//   src : source state
//   dst : dest state
//   v   : input data
// output
//   out : output data (undefined: no output)
//
var invokeScript = []byte(`user := import("user"); out := user[fn](src, dst, immutable(v))`)

// Internal script used to retrieve a function from the user script.
//
// input
//   fn  : function name
// output
//   out : error (or undefined if all validated)
//
var retrieveScript = []byte(`user := import("user"); out := user[fn]`)
