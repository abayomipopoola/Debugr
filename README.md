# Debugr

Debugr is a command-line tool that leverages Claude AI to assist with debugging and testing tasks. It is written using only Go standard libraries, without any additional dependencies.

## Installation

To install Debugr, follow these steps:

1. Ensure you have Go installed on your system (version 1.16 or later).
2. Clone the repository:

   ```
   git clone https://github.com/abayomipopoola/debugr.git
   cd debugr
   ```

3. Install the binary:
   ```
   go install ./cmd/debugr
   ```

This will compile the program and install the binary in your `$GOPATH/bin` directory. Make sure this directory is in your system's PATH.

## Usage

After installation, you can run Debugr from anywhere in your terminal:

```bash
debugr [flags] <prompt>
```

Flags:

- `--context <file>`: Specify a single file as context
- `--context-dir <directory>`: Specify a directory as context
- `--debug`: Enable debug mode
- `--dry-run`: Print suggested actions without executing

### Examples:

#### 1. Write test for this simple addNumber function:

Directory structure:

```bash
$ tree
.
└── adder
   └── adder.go
```

`adder.go`:

```go
package adder

func addNumbers(x, y int) int {
	return x + y
}
```

Before running Debugr, set your ANTHROPIC_API_KEY environment variable:

```bash
export ANTHROPIC_API_KEY=your_api_key_here
```

Then run:

```bash
debugr --debug --context adder/adder.go "write test for the addNumber"
```

Output:

```bash
...
This will run the test and show you the results. If everything is working correctly, you should see a message indicating that the test passed.
To write a test for the `addNumbers` function, we need to create a new test file in the same package. Here's how we can do that:
Execute this action? (y/n): y
File create: adder/adder_test.go
Now you can run the test using the following command:
Execute this action? (y/n): y
ok      test_examples/adder     0.745s
This will run the test and show you the results. If everything is working correctly, you should see a message indicating that the test passed.
```

Updated directory structure:

```bash
$ tree
.
└── adder
   ├── adder.go
   └── adder_test.go
```

`adder_test.go`:

```go
// ...
func TestAddNumbers(t *testing.T) {
	tests := []struct {
		name     string
		x        int
		y        int
		expected int
	}{
		{"positive numbers", 2, 3, 5},
		{"negative numbers", -2, -3, -5},
		{"mixed numbers", -2, 3, 1},
		{"zeros", 0, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := addNumbers(tt.x, tt.y)
			if result != tt.expected {
				t.Errorf("addNumbers(%d, %d) = %d; want %d", tt.x, tt.y, result, tt.expected)
			}
		})
	}
}
```

#### 2. Fixing a Python code bug and testing, with directory context:

Directory structure:

```bash
$ tree
.
├── binarygap
│   └── binary_gap.py
└── twosum
    └── twosum.py
```

`twosum.py` (incorrect solution):

```py
"""
Given an array of integers nums and an integer target, return indices
of the two numbers such that they add up to target.
"""
def solution(nums, target):
    ls = len(nums)
    for i in range(ls):
        for j in range(i, ls):   # bug: should start from i+1
            if nums[i] + nums[j] == target:
                return [i, j]

nums = [3,2,4]
target = 6
print(solution(nums, target))
```

Run:

```bash
debugr --context-dir . "debug the twosum: maintain the inefficient solution for now"
```

Output:

```bash
...
To debug the twosum solution while maintaining its current (inefficient) approach, we can make a few adjustments to improve its correctness and add some debugging information. Here are the steps:
1. Modify the inner loop to start from i+1 to avoid using the same element twice:
Execute this action? (y/n):
```

Updated `twosum.py`:

```py
def solution(nums, target):
    ls = len(nums)
    for i in range(ls):
        for j in range(i + 1, ls ):  # Start from i+1 to avoid using the same element twice
            print(f"Checking: nums[{i}] = {nums[i]}, nums[{j}] = {nums[j]}")  # Debugging print
            if nums[i] + nums[j] == target:
                return [i, j]
    return None  # Return None if no solution is found

# Test cases
nums1 = [3, 2, 4]
target1 = 6
print("Test case 1:", solution(nums1, target1))
...
```

#### 3. Optimizing code and writing tests:

Directory structure:

```bash
$ tree
.
├── binarygap
│   └── binary_gap.py
└── twosum
    └── twosum.py
```

output:

```bash
...
Certainly! I'll optimize the `twosum` function for better time complexity and provide a unit test for it. Here are the steps:
1. Optimize the `twosum` function:
Execute this action? (y/n): y
File modify: twosum/twosum.py
2. Create a new file for unit tests:
Execute this action? (y/n): y
File create: twosum/test_twosum.py
3. Run the unit tests:
The optimized `solution` function now has a time complexity of O(n) instead of O(n^2). It uses a dictionary to store the complement of each number as we iterate through the list, allowing us to find the solution in a single pass.
```

Updated directory structure:

```bash
$ tree
.
├── binarygap
│ └── binary_gap.py
└── twosum
├── test_twosum.py
└── twosum.py

```

Updated `twosum.py`:

```py
def solution(nums, target):
    num_dict = {}
    for i, num in enumerate(nums):
        complement = target - num
        if complement in num_dict:
            return [num_dict[complement], i]
        num_dict[num] = i
    return None  # Return None if no solution is found

# Test cases
if __name__ == "__main__":
    nums1 = [3, 2, 4]
    target1 = 6
    print("Test case 1:", solution(nums1, target1))
...
```

`test_twosum.py`:

```py
import unittest
from twosum import solution

class TestTwoSum(unittest.TestCase):
    def test_basic_case(self):
        self.assertEqual(solution([2, 7, 11, 15], 9), [0, 1])

    def test_same_numbers(self):
        self.assertEqual(solution([3, 3], 6), [0, 1])
    ...

if __name__ == '__main__':
    unittest.main()
```

## TODO

- Implement code auto-formatting (currently leveraging code editor auto-format on save).
- Add support for more languages (currently supports Go, Python, and JavaScript).
