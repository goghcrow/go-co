//go:build co

package linq

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	. "github.com/goghcrow/go-co"
)

type Cons[T1, T2 any] struct {
	Car T1
	Cdr T2
}

func Const[Any, R any](x R) func(Any) R {
	return func(Any) R {
		return x
	}
}

func Id[A any](a A) A {
	return a
}

// tests ref
// https://learn.microsoft.com/en-us/dotnet/api/system.linq.enumerable.aggregate?view=net-8.0

type (
	Signed interface {
		~int | ~int8 | ~int16 | ~int32 | ~int64
	}
	Unsigned interface {
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
	}
	Integer interface{ Signed | Unsigned }
	Float   interface{ ~float32 | ~float64 }
	Number  interface{ Integer | Float }
)

func eq[A comparable](n A) Predicate[A] { return func(x A) bool { return x == n } }
func lt[N Number](n N) Predicate[N]     { return func(x N) bool { return x < n } }
func gt[N Number](n N) Predicate[N]     { return func(x N) bool { return x > n } }
func le[N Number](n N) Predicate[N]     { return func(x N) bool { return x <= n } }
func ge[N Number](n N) Predicate[N]     { return func(x N) bool { return x >= n } }

func startsWith(prefix string) Predicate[string] {
	return func(s string) bool {
		return strings.HasPrefix(s, prefix)
	}
}

func isEven(x int) bool { return x%2 == 0 }
func square(x int) int  { return x * x }
func double(x int) int  { return x * 2 }

func assertEqual(t *testing.T, x, y any) {
	if !reflect.DeepEqual(x, y) {
		t.Fail()
	}
}

func ToSlice[A any](it Iter[A]) (xs []A) {
	for a := range it {
		xs = append(xs, a)
	}
	return
}

func TestDeferred(t *testing.T) {
	it := func() (_ Iter[int]) {
		panic("deferred")
		Yield(0)
		return
	}
	Select[int](it(), double)
}

func TestOnce(t *testing.T) {
	xs := Of(1, 2, 3)
	ys := Select(xs, Id[int])
	assertEqual(t, ToSlice(ys), []int{1, 2, 3})

	zs := Select(xs, Id[int])
	assertEqual(t, len(ToSlice(zs)), 0)
}

func TestOnce1(t *testing.T) {
	type T = Cons[string, int]
	xs := Of("a", "b", "c")
	ys := Of(1, 2, 3)
	zs := SelectMany(xs, func(x string) Iter[T] {
		return SelectMany(ys, func(y int) Iter[T] {
			return Unit(T{x, y})
		})
	})
	assertEqual(t, ToSlice(zs), []T{
		{"a", 1}, T{"a", 2}, T{"a", 3},
	})
}

func TestSelect(t *testing.T) {
	xs := Of(1, 2, 3)
	ys := Select(xs, square)
	assertEqual(t, ToSlice(ys), []int{1, 4, 9})
}

func TestWhere(t *testing.T) {
	xs := Range(1, 10)
	ys := Where(xs, isEven)
	assertEqual(t, ToSlice(ys), []int{2, 4, 6, 8})
}

func TestSkip(t *testing.T) {
	xs := Range(1, 10)
	ys := Skip(
		Where(xs, isEven),
		2,
	)
	assertEqual(t, ToSlice(ys), []int{6, 8})
}

func TestSkipWhile(t *testing.T) {
	xs := Range(1, 10)
	ys := SkipWhile(xs, lt(5))
	assertEqual(t, ToSlice(ys), []int{5, 6, 7, 8, 9})
}

func TestTake(t *testing.T) {
	xs := Range(1, 10)
	ys := Take(
		Where(xs, isEven),
		3,
	)
	assertEqual(t, ToSlice(ys), []int{2, 4, 6})
}

func TestTakeWhile(t *testing.T) {
	xs := Range(1, 10)
	ys := TakeWhile(xs, lt(5))
	assertEqual(t, ToSlice(ys), []int{1, 2, 3, 4})
}

func TestFirst(t *testing.T) {
	xs := Of(9, 34, 65, 92, 87, 435, 3, 54,
		83, 23, 87, 435, 67, 12, 19)
	first, ok := First(xs)
	assertEqual(t, first, 9)
	assertEqual(t, ok, true)
}

func TestFirstWhile(t *testing.T) {
	xs := Of(9, 34, 65, 92, 87, 435, 3, 54,
		83, 23, 87, 435, 67, 12, 19)
	first, ok := FirstWhile(xs, gt(80))
	assertEqual(t, first, 92)
	assertEqual(t, ok, true)
}

func TestLast(t *testing.T) {
	xs := Of(9, 34, 65, 92, 87, 435, 3, 54,
		83, 23, 87, 67, 12, 19)
	first, ok := Last(xs)
	assertEqual(t, first, 19)
	assertEqual(t, ok, true)
}

func TestLastWhile(t *testing.T) {
	xs := Of(9, 34, 65, 92, 87, 435, 3, 54,
		83, 23, 87, 67, 12, 19)
	first, ok := LastWhile(xs, gt(80))
	assertEqual(t, first, 87)
	assertEqual(t, ok, true)
}

func TestAggregate(t *testing.T) {
	fruits := Of("apple", "mango", "orange", "passionfruit", "grape")
	s := Aggregate[string, string](fruits, "banana", func(longest string, next string) string {
		if len(next) > len(longest) {
			return next
		}
		return longest
	}, strings.ToUpper)

	assertEqual(t, s, "PASSIONFRUIT")
}

func TestFold(t *testing.T) {
	{
		xs := Of(1, 4, 5)
		r := Fold(xs, 5, func(acc int, cur int) int {
			return acc*2 + cur
		})
		assertEqual(t, r, 57)
	}

	{
		xs := Of(4, 8, 8, 3, 9, 0, 7, 8, 2)
		r := Fold(xs, 0, func(total int, cur int) int {
			if isEven(cur) {
				return total + 1
			}
			return total
		})
		assertEqual(t, r, 6)
	}
}

func TestReduce(t *testing.T) {
	sentence := "the quick brown fox jumps over the lazy dog"
	words := strings.Split(sentence, " ")
	reversed, ok := Reduce(Of(words...), func(acc string, cur string) string {
		return cur + " " + acc
	})
	assertEqual(t, ok, true)
	assertEqual(t, reversed, "dog lazy the over jumps fox brown quick the")
}

func TestAll(t *testing.T) {
	{
		pets := []string{"Barley", "Boots", "Whiskers"}
		allStartWithB := All(Of(pets...), startsWith("B"))
		assertEqual(t, allStartWithB, false)
	}
	{
		pets := []string{"Barley", "Boots"}
		allStartWithB := All(Of(pets...), startsWith("B"))
		assertEqual(t, allStartWithB, true)
	}
}

func TestAllEx(t *testing.T) {
	type Pet struct {
		Name string
		Age  int
	}
	type Person struct {
		LastName string
		Pets     []Pet
	}

	people := []Person{
		{
			LastName: "Haas",
			Pets: []Pet{
				{"Barley", 10},
				{"Boots", 14},
				{"Whiskers", 6},
			},
		},
		{
			LastName: "Fakhouri",
			Pets: []Pet{
				{"Snowball", 1},
			},
		},
		{
			LastName: "Antebi",
			Pets: []Pet{
				{"Belle", 8},
			},
		},
		{
			LastName: "Philips",
			Pets: []Pet{
				{"Sweetie", 2},
				{"Rover", 13},
			},
		},
	}

	petAgeGt5 := func(pet Pet) bool {
		return pet.Age > 5
	}
	personLastName := func(person Person) string {
		return person.LastName
	}
	// Determine which people have pets that are all older than 5.
	// IEnumerable<string> names = from person in people
	//                             where person.Pets.All(pet => pet.Age > 5)
	//                             select person.LastName;
	names := Select(
		Where(
			Of(people...),
			func(person Person) bool {
				return All(
					Of(person.Pets...),
					petAgeGt5,
				)
			},
		),
		personLastName,
	)

	assertEqual(t, ToSlice(names), []string{
		"Haas",
		"Antebi",
	})
}

func TestAnyElem(t *testing.T) {
	xs := Of(1, 2)
	assertEqual(t, AnyElem(xs), true)
}

func TestAnyElemEx(t *testing.T) {
	type Pet struct {
		Name string
		Age  int
	}
	type Person struct {
		LastName string
		Pets     []Pet
	}

	people := []Person{
		{
			LastName: "Haas",
			Pets: []Pet{
				{"Barley", 10},
				{"Boots", 14},
				{"Whiskers", 6},
			},
		},
		{
			LastName: "Fakhouri",
			Pets: []Pet{
				{"Snowball", 1},
			},
		},
		{
			LastName: "Antebi",
			Pets:     []Pet{
				// {"Belle", 8},
			},
		},
		{
			LastName: "Philips",
			Pets: []Pet{
				{"Sweetie", 2},
				{"Rover", 13},
			},
		},
	}

	personLastName := func(person Person) string {
		return person.LastName
	}

	// Determine which people have a non-empty Pet array.
	// IEnumerable<string> names = from person in people
	//                             where person.Pets.Any()
	//                             select person.LastName;
	names := Select(
		Where(
			Of(people...),
			func(person Person) bool {
				return AnyElem(
					Of(person.Pets...),
				)
			},
		),
		personLastName,
	)

	assertEqual(t, ToSlice(names), []string{
		"Haas",
		"Fakhouri",
		"Philips",
	})
}

func TestAny(t *testing.T) {
	type Pet struct {
		Name       string
		Age        int
		Vaccinated bool
	}

	//    // Determine whether any pets over age 1 are also unvaccinated.
	//    bool unvaccinated = pets.Any(p => p.Age > 1 && p.Vaccinated == false)
	pets := []Pet{
		{"Barley", 8, true},
		{"Boots", 4, false},
		{"Whiskers", 1, false},
	}
	unvaccinated := Any(
		Of(pets...),
		func(pet Pet) bool {
			return pet.Age > 1 && !pet.Vaccinated
		},
	)
	assertEqual(t, unvaccinated, true)
}

func TestAppend(t *testing.T) {
	xs := []int{1, 2, 3, 4}

	ys := Append(Of(xs...), 5)

	assertEqual(t,
		strings.Join(ToSlice(Select(Of(xs...), strconv.Itoa)), ", "),
		"1, 2, 3, 4",
	)
	assertEqual(t,
		strings.Join(ToSlice(Select(ys, strconv.Itoa)), ", "),
		"1, 2, 3, 4, 5",
	)
}

// left outer join
func TestCrossJoin(t *testing.T) {
	// from inner in items
	// from outer in function(items)
	// select projection(inner, outer)

	type T = Cons[string, int]

	xs := []string{"a", "b", "c"}
	ys := []int{1, 2, 3}

	zs := SelectMany(Of(xs...), func(x string) Iter[T] {
		return SelectMany(Of(ys...), func(y int) Iter[T] {
			return Unit(T{x, y})
		})
	})

	assertEqual(t, ToSlice(zs), []T{
		{"a", 1}, T{"a", 2}, T{"a", 3},
		{"b", 1}, T{"b", 2}, T{"b", 3},
		{"c", 1}, T{"c", 2}, T{"c", 3},
	})
}
