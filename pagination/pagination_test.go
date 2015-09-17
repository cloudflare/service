package pagination

import (
	"testing"
)

func TestPageCount(t *testing.T) {

	var (
		total        int64
		limit        int64
		correctValue int64
		result       int64
	)

	message := "PageCount(%d, %d) = %d should be %d"

	// limit = 25 (default)
	limit = DefaultLimit

	total = 0
	correctValue = 0
	result = PageCount(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}

	total = 1
	correctValue = 1
	result = PageCount(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}

	total = 24
	correctValue = 1
	result = PageCount(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}

	total = 25
	correctValue = 1
	result = PageCount(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}

	total = 26
	correctValue = 2
	result = PageCount(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}

	total = 49
	correctValue = 2
	result = PageCount(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}

	total = 50
	correctValue = 2
	result = PageCount(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}

	total = 51
	correctValue = 3
	result = PageCount(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}

	// limit = 5
	limit = 5

	total = 0
	correctValue = 0
	result = PageCount(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}

	total = 1
	correctValue = 1
	result = PageCount(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}

	total = 4
	correctValue = 1
	result = PageCount(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}

	total = 5
	correctValue = 1
	result = PageCount(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}

	total = 6
	correctValue = 2
	result = PageCount(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}

	total = 9
	correctValue = 2
	result = PageCount(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}

	total = 10
	correctValue = 2
	result = PageCount(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}

	total = 11
	correctValue = 3
	result = PageCount(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}
}

func TestMaxOffset(t *testing.T) {

	var (
		total        int64
		limit        int64
		correctValue int64
		result       int64
	)

	message := "MaxOffset(%d, %d) = %d should be %d"

	// limit = 25 (default)
	limit = DefaultLimit

	total = 0
	correctValue = 0
	result = MaxOffset(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}

	total = 24
	correctValue = 0
	result = MaxOffset(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}

	total = 25
	correctValue = 0
	result = MaxOffset(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}

	total = 26
	correctValue = 25
	result = MaxOffset(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}

	total = 49
	correctValue = 25
	result = MaxOffset(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}

	total = 50
	correctValue = 25
	result = MaxOffset(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}

	total = 51
	correctValue = 50
	result = MaxOffset(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}

	// limit = 25 (default)
	limit = 5

	total = 0
	correctValue = 0
	result = MaxOffset(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}

	total = 4
	correctValue = 0
	result = MaxOffset(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}

	total = 5
	correctValue = 0
	result = MaxOffset(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}

	total = 6
	correctValue = 5
	result = MaxOffset(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}

	total = 9
	correctValue = 5
	result = MaxOffset(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}

	total = 10
	correctValue = 5
	result = MaxOffset(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}

	total = 11
	correctValue = 10
	result = MaxOffset(total, limit)
	if result != correctValue {
		t.Errorf(message, total, limit, result, correctValue)
	}
}

func TestOffsetFromPage(t *testing.T) {
	var (
		page           int64
		limit          int64
		expectedOffset int64
		result         int64
	)

	message := "OffsetFromPage(%d, %d) = %d should be %d"

	// limit = 25 (default)
	limit = DefaultLimit

	page = 0
	expectedOffset = 0
	result = OffsetFromPage(page, limit)
	if result != expectedOffset {
		t.Errorf(message, page, limit, result, expectedOffset)
	}

	page = 1
	expectedOffset = 0
	result = OffsetFromPage(page, limit)
	if result != expectedOffset {
		t.Errorf(message, page, limit, result, expectedOffset)
	}

	page = 2
	expectedOffset = 25
	result = OffsetFromPage(page, limit)
	if result != expectedOffset {
		t.Errorf(message, page, limit, result, expectedOffset)
	}

	page = 3
	expectedOffset = 50
	result = OffsetFromPage(page, limit)
	if result != expectedOffset {
		t.Errorf(message, page, limit, result, expectedOffset)
	}

	// Change limit to 20
	limit = 20

	page = 1
	expectedOffset = 0
	result = OffsetFromPage(page, limit)
	if result != expectedOffset {
		t.Errorf(message, page, limit, result, expectedOffset)
	}

	page = 2
	expectedOffset = 20
	result = OffsetFromPage(page, limit)
	if result != expectedOffset {
		t.Errorf(message, page, limit, result, expectedOffset)
	}

	page = 3
	expectedOffset = 40
	result = OffsetFromPage(page, limit)
	if result != expectedOffset {
		t.Errorf(message, page, limit, result, expectedOffset)
	}

	page = 4
	expectedOffset = 60
	result = OffsetFromPage(page, limit)
	if result != expectedOffset {
		t.Errorf(message, page, limit, result, expectedOffset)
	}
}
