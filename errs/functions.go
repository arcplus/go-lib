package errs

import "github.com/juju/errors"

// Trace adds the location of the Trace call to the stack.  The Cause of the
// resulting error is the same as the error parameter.  If the other error is
// nil, the result will be nil.
//
// For example:
//   if err := SomeFunc(); err != nil {
//       return errors.Trace(err)
//   }
//
var Trace = errors.Trace

// Annotate is used to add extra context to an existing error. The location of
// the Annotate call is recorded with the annotations. The file, line and
// function are also recorded.
//
// For example:
//   if err := SomeFunc(); err != nil {
//       return errors.Annotate(err, "failed to frombulate")
//   }
//
var Annotate = errors.Annotate

// Annotatef is used to add extra context to an existing error. The location of
// the Annotate call is recorded with the annotations. The file, line and
// function are also recorded.
//
// For example:
//   if err := SomeFunc(); err != nil {
//       return errors.Annotatef(err, "failed to frombulate the %s", arg)
//   }
//
var Annotatef = errors.Annotatef

// DeferredAnnotatef annotates the given error (when it is not nil) with the given
// format string and arguments (like fmt.Sprintf). If *err is nil, DeferredAnnotatef
// does nothing. This method is used in a defer statement in order to annotate any
// resulting error with the same message.
//
// For example:
//
//    defer DeferredAnnotatef(&err, "failed to frombulate the %s", arg)
//
var DeferredAnnotatef = errors.DeferredAnnotatef

// Wrap changes the Cause of the error. The location of the Wrap call is also
// stored in the error stack.
//
// For example:
//   if err := SomeFunc(); err != nil {
//       newErr := &packageError{"more context", private_value}
//       return errors.Wrap(err, newErr)
//   }
//
var Wrap = errors.Wrap

// Wrapf changes the Cause of the error, and adds an annotation. The location
// of the Wrap call is also stored in the error stack.
//
// For example:
//   if err := SomeFunc(); err != nil {
//       return errors.Wrapf(err, simpleErrorType, "invalid value %q", value)
//   }
//
var Wrapf = errors.Wrapf

// Mask hides the underlying error type, and records the location of the masking.
var Mask = errors.Mask

// Mask masks the given error with the given format string and arguments (like
// fmt.Sprintf), returning a new error that maintains the error stack, but
// hides the underlying error type.  The error string still contains the full
// annotations. If you want to hide the annotations, call Wrap.
var Maskf = errors.Maskf

// Cause returns the cause of the given error.  This will be either the
// original error, or the result of a Wrap or Mask call.
//
// Cause is the usual way to diagnose errors that may have been wrapped by
// the other errors functions.
var Cause = errors.Cause

// ErrorStack returns a string representation of the annotated error. If the
// error passed as the parameter is not an annotated error, the result is
// simply the result of the Error() method on that error.
//
// If the error is an annotated error, a multi-line string is returned where
// each line represents one entry in the annotation stack. The full filename
// from the call stack is used in the output.
//
//     first error
//     github.com/juju/errors/annotation_test.go:193:
//     github.com/juju/errors/annotation_test.go:194: annotation
//     github.com/juju/errors/annotation_test.go:195:
//     github.com/juju/errors/annotation_test.go:196: more context
//     github.com/juju/errors/annotation_test.go:197:
var ErrorStack = errors.ErrorStack
