package main

import (
	"bufio"
	"io"
	"strings"
)

type segmentsScanner struct {
	scanner                 *bufio.Scanner
	state                   int
	targetSegment           string
	isTargetToTargetSegment bool
	isTargetSegmentFound    bool
	segmentIndentation      string
	segmentBuilder          strings.Builder
	done                    bool
	err                     error
}

func NewSegmentsScanner(reader io.Reader, targetSegment string) *segmentsScanner {
	return &segmentsScanner{
		scanner:       bufio.NewScanner(reader),
		targetSegment: targetSegment,
	}
}

func (r *segmentsScanner) setToken(line string) {
	r.segmentBuilder.WriteString(line)
	r.segmentBuilder.WriteByte('\n')
}

func (r *segmentsScanner) isCollectable() bool {
	return r.isTargetSegmentFound || r.isTargetToTargetSegment
}

func (r *segmentsScanner) setSegmentToken(line string) {
	if r.isCollectable() {
		r.setToken(strings.TrimPrefix(line, r.segmentIndentation))
	}
}

func (r *segmentsScanner) tryStartSegment(line string) bool {
	matched := SEGMENT_NAME_REG_EXP.FindStringSubmatch(line)
	if matched == nil {
		return false
	}
	r.state = SEGMENT_STARTS
	if r.targetSegment == matched[1] {
		r.isTargetSegmentFound = true
	} else if r.targetSegment == DEFAULT_TARGET_SEGMENT ||
		(len(matched) > 2 && strings.Contains(matched[2], r.targetSegment)) {
		r.isTargetToTargetSegment = true
	}
	return true
}

func (r *segmentsScanner) finishSegment(line string) bool {
	if r.isTargetSegmentFound {
		r.state = TARGET_SEGMENT_FINISHED
		r.done = true
		return true
	}
	wasCollectable := r.isTargetToTargetSegment
	r.isTargetToTargetSegment = false
	if r.tryStartSegment(line) {
		return wasCollectable || r.isCollectable()
	}
	r.state = SEGMENT_NOT_DEFINED
	r.setToken(line)
	return wasCollectable
}

func (r *segmentsScanner) processLine(line string) bool {
	switch r.state {
	case SEGMENT_NOT_DEFINED:
		if r.tryStartSegment(line) {
			return r.isCollectable()
		} else {
			r.setToken(line)
		}
	case SEGMENT_STARTS:
		matches := SEGMENT_INDENT_REG_EXP.FindStringSubmatch(line)
		if matches != nil {
			r.segmentIndentation = matches[1]
			r.state = SEGMENT_CONTINUED
			r.setSegmentToken(line)
		} else {
			return r.finishSegment(line)
		}
	case SEGMENT_CONTINUED:
		if strings.HasPrefix(line, r.segmentIndentation) {
			r.setSegmentToken(line)
		} else {
			return r.finishSegment(line)
		}
	}
	return false
}

func (r *segmentsScanner) Scan() bool {
	if r.done {
		return false
	}
	r.segmentBuilder.Reset()
	isSegmentUpdated := false
	for !isSegmentUpdated && r.scanner.Scan() {
		isSegmentUpdated = r.processLine(r.scanner.Text())
	}
	if isSegmentUpdated {
		return !r.done
	}
	if err := r.scanner.Err(); err != nil {
		r.err = err
	} else if !r.isTargetSegmentFound && r.targetSegment != DEFAULT_TARGET_SEGMENT {
		r.err = ErrSegmentNotFound
	}
	r.done = true
	return false
}

func (r *segmentsScanner) Err() error {
	return r.err
}

func (r *segmentsScanner) Text() string {
	return r.segmentBuilder.String()
}
