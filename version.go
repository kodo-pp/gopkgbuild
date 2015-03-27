package pkgbuild

import "strconv"

// Version string
type Version string

type CompleteVersion struct {
	Version Version
	Epoch   int
	Pkgrel  int
}

// Compare alpha and numeric segments of two versions.
// return 1: a is newer than b
//        0: a and b are the same version
//       -1: b is newer than a
//
// This is based on the rpmvercmp function used in libalpm
// https://projects.archlinux.org/pacman.git/tree/lib/libalpm/version.c
func rpmvercmp(a, b Version) int {
	if a == b {
		return 0
	}

	var one, two, ptr1, ptr2 int
	var isNum bool
	one, two, ptr1, ptr2 = 0, 0, 0, 0

	// loop through each version segment of a and b and compare them
	for len(a) > one && len(b) > two {
		for len(a) > one && !isAlphaNumeric(a[one]) {
			one++
		}
		for len(b) > two && !isAlphaNumeric(b[two]) {
			two++
		}

		// if we ran to the end of either, we are finished with the loop
		if !(len(a) > one && len(b) > two) {
			break
		}

		// if the seperator lengths were different, we are also finished
		if one-ptr1 != two-ptr2 {
			if one-ptr1 < two-ptr2 {
				return -1
			}
			return 1
		}

		ptr1 = one
		ptr2 = two

		// grab first completely alpha or completely numeric segment
		// leave one and two pointing to the start of the alpha or numeric
		// segment and walk ptr1 and ptr2 to end of segment
		if isDigit(a[ptr1]) {
			for len(a) > ptr1 && isDigit(a[ptr1]) {
				ptr1++
			}
			for len(b) > ptr2 && isDigit(b[ptr2]) {
				ptr2++
			}
			isNum = true
		} else {
			for len(a) > ptr1 && isAlpha(a[ptr1]) {
				ptr1++
			}
			for len(b) > ptr2 && isAlpha(b[ptr2]) {
				ptr2++
			}
			isNum = false
		}

		// take care of the case where the two version segments are
		// different types: one numeric, the other alpha (i.e. empty)
		// numeric segments are always newer than alpha segments
		if two == ptr2 {
			if isNum {
				return 1
			}
			return -1
		}

		if isNum {
			// we know this part of the strings only contains digits
			// so we can ignore the error value since it should
			// always be nil
			as, _ := strconv.ParseInt(string(a[one:ptr1]), 10, 0)
			bs, _ := strconv.ParseInt(string(b[two:ptr2]), 10, 0)

			// whichever number has more digits wins
			if as > bs {
				return 1
			}
			if as < bs {
				return -1
			}
		} else {
			cmp := alphaCompare(a[one:ptr1], b[two:ptr2])
			if cmp < 0 {
				return -1
			}
			if cmp > 0 {
				return 1
			}
		}

		// advance one and two to next segment
		one = ptr1
		two = ptr2
	}

	// this catches the case where all numeric and alpha segments have
	// compared identically but the segment separating characters were
	// different
	if len(a) <= one && len(b) <= two {
		return 0
	}

	// the final showdown. we never want a remaining alpha string to
	// beat an empty string. the logic is a bit weird, but:
	// - if one is empty and two is not an alpha, two is newer.
	// - if one is an alpha, two is newer.
	// - otherwise one is newer.
	if (len(a) <= one && !isAlpha(b[two])) || len(a) > one && isAlpha(a[one]) {
		return -1
	}
	return 1
}

// alphaCompare compares two alpha version segments and will return a positive
// value if a is bigger than b and a negative if b is bigger than a else 0
func alphaCompare(a, b Version) int8 {
	if a == b {
		return 0
	}

	i := 0
	for len(a) > i && len(b) > i && a[i] == b[i] {
		i++
	}

	if len(a) == i && len(b) > i {
		return -1
	}

	if len(b) == i {
		return 1
	}

	return int8(a[i]) - int8(b[i])
}

// check if version number v is bigger than v2
func (v Version) bigger(v2 Version) bool {
	return rpmvercmp(v, v2) == 1
}

// isAlphaNumeric reports whether c is an alpha character or digit
func isAlphaNumeric(c uint8) bool {
	return isDigit(c) || isAlpha(c)
}

// isAlpha reports whether c is an alpha character
func isAlpha(c uint8) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}

// isDigit reports whether d is an ASCII digit
func isDigit(d uint8) bool {
	return '0' <= d && d <= '9'
}
