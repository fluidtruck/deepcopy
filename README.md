# DeepCopy

DeepCopy is a Go library for recursively copying an object
into another object.

It is designed to replace manual conversion between models
that Go deems incompatible but have underlying, compatible
field types.

DeepCopy allows for extraordinary flexibility in converting
between different structs and other types. It performs automatic type casting
for all fields that typically require manual conversion, such
as between uint64 and uint or int and string. In addition, it automatically converts
between pointers and non-pointers at any level (e.g. **string to string and vice versa). It can handle
slices, maps, nested structs, time.Time objects,
protobuf.timestamppb objects, and more. It additionally supports
an optional tag used to manually set field names for more
directed field matching.
## Table of Contents
* [How to Install](#how-to-install)
* [How to Use DeepCopy](#how-to-use-deepcopy)
    * [Case 1: Struct Conversion](#case-1-struct-conversion)
    * [Case 2: Identical Copy](#case-2-identical-copy)
    * [Case 3: General Type Casting](#case-3-general-type-casting)
* [What Gets Copied?](#what-exactly-gets-copied?)
* Examples
    * [Basic Example](#basic-example)
    * [Pointers](#pointers)
    * [Errors](#errors)

## How to Install
From the command line:
```go 
go get github.com/fluidtruck/deepcopy
```
At the top of your file:
```go 
import "github.com/fluidtruck/deepcopy"
```

## How To Use DeepCopy
DeepCopy has 3 main use cases:
1. DeepCopy can convert objects into different struct
   types by copying over (and, where necessary, automatically type casting)
   all matching fields between two structs.
2. DeepCopy can also create a recursive identical
   copy of an object (a true deep copy).
3. Finally, DeepCopy can be used as a general, all-purpose type casting
   function. It does not require that the target type be known in order to work.

### Case 1: Struct Conversion
Let ```objA``` be an object of type ```StructA```. This is the object
that we want copied.\
Let ```objB``` be an object of type ```StructB```. This is the object
that we want to copy into. \
\
Call DeepCopy by passing ```objA``` and a pointer to ```objB``` as arguments.
```go
err := deepCopy.DeepCopy(objA, &objB)
```
> Note: The second argument to DeepCopy must ***always*** be a pointer.
> Otherwise, an [error](#expected-pointer) will be returned.

Done! Now, ```objB```has all of the field values of ```objA```
recursively copied over. \
\
See [what exactly](#what-gets-copied) gets copied.

### Case 2: Identical Copy
Let ```objA``` be an object of type ```A```. This is the object
that we want to copy. \
\
First, create a new, empty object of type A.
```go
copyA := A{}
```
Next, call DeepCopy by passing ```objA``` and a pointer to ```copyA```
as arguments.
```go
err := deepCopy.DeepCopy(objA, &copyA)
```
> Note: The second argument to DeepCopy must ***always*** be a pointer.
> Otherwise, an [error](#expected-pointer) will be returned.

Done! Now  ```copyA``` is an exactly identical copy of ```objA```.

### Case 3: General Type Casting
Let's say that we want to copy ```objA``` into ```objB```, but we're not sure what objB is.
```go
objA := 4
// objB = ? ? ? ? ?
```
Call DeepCopy by passing ```objA``` and a pointer to ```objB``` as arguments.
```go
err := deepCopy.DeepCopy(objA, &objB)
```
> Note: The second argument to DeepCopy must ***always*** be a pointer.
> Otherwise, an [error](#expected-pointer) will be returned.

Done! Now, ```objB``` will have an equivalent value to ```objA```. \
If ```objB``` was a string, then ```objB``` will have value ```"4"```. \
If ```objB``` was a float32, then ```objB``` will have value ```float32(4)```. \
If ```objB``` was a uint64, then ```objB``` will have value ```uint64(4)```. \
...etc.

## What exactly gets copied?
Let ```objA``` be an object of type ```StructA```. \
Let ```objB``` be an object of type ```StructB```.
```go
err := deepcopy.DeepCopy(objA, &objB)
```

Assuming that ```objA``` is a struct, then all fields of ```objA``` that
* (1) are not null,

and
* (2) [match](#matching-fields) a field in ```Struct B```

will be copied over to the matching field in ```objB```.

Additionally, all existing fields in ```objB``` that are ***not
overwritten*** by ```objA``` will remain in ```objB```.

### Matching Fields
Fields are considered matching if they have the same name (case-insensitive)
or if one field's name matches another field's "dc" tag. \
\
Field matches can be manually set by using the "dc" tag.\
\
In the following example, all fields in ```StructA``` are considered
to have a respective matching field in ```StructB```.
```go
type StructA struct {
   FieldOne string // matches with StructB.FieldOne
   TheSecondField uint // matches with StructB.Thesecondfield
   FieldThree time.Time // matches with StructB.FieldThreeAlternativeName because of dc tag
   FieldFour **string `dc:"field4"` // matches with StructB.Field4 because of dc tag
   FIELDFIVE *bool// matches with StructB.Fieldfive
}

type StructB struct {
   FieldOne string
   Thesecondfield uint32
   FieldThreeAlternativeName time.Time `dc:"fieldthree"`
   Field4 *int32
   FieldFive ***bool
}
```
If an object of type ```StructA``` and a pointer to an object of type ```StructB```
are passed into DeepCopy, then DeepCopy will attempt to copy all non-null ```StructA```
fields into the object of type ```StructB```.

Field types are not considered when determining whether two fields match.
If a non-null ```StructA``` field has a matching ```StructB``` field whose
type is incompatible with the original ```StructA```field's type (for example,
a string array and a time.Time pointer), then DeepCopy will throw an Unable
to Convert [error](#unable-to-convert), such as in the case below:
```go
import (
    dc "github.com/fluidtruck/deepcopy"
)

type StructA struct {
    Foo uint64
}

type StructB struct {
    Foo bool
}

func main() {
    a := StructA{Foo: uint64(12)}
    b := StructB{}
    err := dc.DeepCopy(a, &b)
    if err ! = nil {
        fmt.Println(err) // will print Err Could Not Convert
    }
}
```

All unexported fields (starting with a lowercase letter) are not considered by
DeepCopy and will not be copied.

## Examples
### Basic Example
```go 
import (
    dc "github.com/fluidtruck/deepcopy"
)

type StructA struct {
    Foo uint64
}

type StructB struct {
    Foo uint
}

func main() {
    a := StructA{Foo: uint64(12)}
    b := StructB{}
    err := dc.DeepCopy(a, &b)
    if err ! = nil {
        fmt.Println(err)
    }
    
    fmt.Printf("%T", b.Foo) // uint
    fmt.Println(b.Foo) // 12
}
```

### Pointers

DeepCopy can reference or dereference values as many times as needed to
convert from the source field type to the target field type.

```go 
import (
    dc "github.com/fluidtruck/deepcopy"
)

type StructA struct {
    Foo uint64
    Bar uint64
    Zak bool
}

type StructB struct {
    Foo *int32
    Bar **string
    Zak ***bool
}

func main() {
    a := StructA{
        Foo: uint64(12),
        Bar: uint64(13),
        Zak: true,
    }
    b := StructB{}
    err := dc.DeepCopy(a, &b)
    if err ! = nil {
        fmt.Println(err)
    }
    
    fmt.Printf("%T", b.Foo) // uint
    fmt.Println(*b.Foo) // 12
    
    fmt.Printf("%T", b.Bar) // **string
    fmt.Println(**b.Bar) // "13"
    
    fmt.Printf("%T", b.Zak) // ***bool
    fmt.Println(***b.Zak) // true
}
```
DeepCopy is bi-directional, so this conversion also works in
reverse.
```go
import (
    dc "github.com/fluidtruck/deepcopy"
)

type StructA struct {
    Foo uint64
    Bar uint64
    Zak bool
}

type StructB struct {
    Foo *int32
    Bar **string
    Zak ***bool
}

var (
    foo = int32(17)
   
    bar = "bar"
    barAddr = &bar
   
    zak = true
    zakAddr = &zak
    zakAddrAddr = &zakAddr
)

func main() {
   b := StructB{
      Foo: &foo,
      Bar: &barAddr,
      zak: &zakAddrAddr
    }
    a := StructA{}
    err := dc.DeepCopy(b, &a)
    if err != nil {
       fmt.Println(err)
    }
    
    fmt.Printf("%T", a.Foo) // uint64
    fmt.Println(a.Foo) // 12
    
    fmt.Printf("%T", b.Bar) // uint64
    fmt.Println(b.Bar) // "13"
    
    fmt.Printf("%T", b.Zak) // bool
    fmt.Println(b.Zak) // true
}
```
Take a look at the test files for more specific examples.


### Errors

> ##### Expected Pointer
> ##### Error: expected pointer for arg1...but received...
> This error occurs when the second argument to DeepCopy() is not a pointer. \
> \
> For example, given
> ```go
> objA := StructA{Foo: bar}
> objB := StructB{}
> err := deepcopy.DeepCopy(objA, objB)
> ```
> The last line should be rewritten as
> ```go
> err := deepcopy.DeepCopy(objA, &objB)
> ```

> ##### Unable to Convert
> ##### Error: unable to convert objA (type ObjAType) to type ObjBType
> This error occurs when DeepCopy is attempting a conversion between two types
> that are incompatible.\
> \
> For example, attempting to convert an int32 object to a time.Time will result
> in this error. \
> \
> Be aware of field name [matches](#matching-fields).
