package distributor

import "sort"

type TestFile struct {
	Path  string
	Suite string
}

type WorkerBucket struct {
	WorkerID int
	Tests    []TestFile
}

type Distribution struct {
	Workers []WorkerBucket
}

func RoundRobin(tests []TestFile, workerCount int) Distribution {
	if workerCount <= 0 {
		workerCount = 1
	}

	buckets := make([]WorkerBucket, workerCount)
	for i := range buckets {
		buckets[i] = WorkerBucket{
			WorkerID: i,
			Tests:    []TestFile{},
		}
	}

	for i, test := range tests {
		workerIdx := i % workerCount
		buckets[workerIdx].Tests = append(buckets[workerIdx].Tests, test)
	}

	return Distribution{Workers: buckets}
}

func (d Distribution) TestCount() int {
	count := 0
	for _, w := range d.Workers {
		count += len(w.Tests)
	}
	return count
}

func (d Distribution) WorkerCount() int {
	return len(d.Workers)
}

func (d Distribution) GetWorkerTests(workerID int) []TestFile {
	if workerID < 0 || workerID >= len(d.Workers) {
		return nil
	}
	return d.Workers[workerID].Tests
}

// Shard returns the slice of tests assigned to shardIndex (1-indexed) of shardTotal.
// Tests are sorted by path for determinism, then file i goes to shard (i % shardTotal) + 1,
// so the combination of all shards covers every test exactly once with no overlap.
func Shard(tests []TestFile, shardIndex, shardTotal int) []TestFile {
	if shardTotal <= 1 {
		return tests
	}
	sorted := make([]TestFile, len(tests))
	copy(sorted, tests)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Path < sorted[j].Path })

	sliced := make([]TestFile, 0, (len(sorted)+shardTotal-1)/shardTotal)
	for i, t := range sorted {
		if i%shardTotal == shardIndex-1 {
			sliced = append(sliced, t)
		}
	}
	return sliced
}
