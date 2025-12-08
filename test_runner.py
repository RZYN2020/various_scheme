import subprocess
import argparse

INTERPRETER_CONFIG = {
    "python": {
        "compile_command": None,
        "run_command": "uv run ./ss_py/scheme.py"
    },
    "c": {
        "compile_command": "make -f ./ss_c/Makefile",
        "run_command": "./ss_c/cscheme"
    },
    "typescript": {
        "compile_command": "cd ./ss_ts && npx tsc",
        "run_command": "node ./ss_ts/dist/scheme.js"
    },
    "go": {
        "compile_command": None,
        "run_command": "go run ./ss_go/scheme.go"
    }

}

TEST_CONFIG = {
    "basic_arithmetic": "tests/basic_arithmetic.ss",
    "conditional": "tests/conditional.ss",
    "lambda": "tests/lambda.ss",
    "recursion": "tests/recursion.ss",
}


class TestCase:
    def __init__(self, expression: str, expected: str):
        self.expression = expression
        self.expected = expected

def load_tests_from_file(file_path) -> list[TestCase]:
    with open(file_path, 'r') as f:
        tests = []
        last_line = None
        for line in f:
            line = line.strip()
            if line.startswith("; expected:"):
                expected = line.split(":")[1].strip()
                assert last_line is not None, "Expected an expression before expected result"
                tests.append(TestCase(last_line, expected))
                last_line = None
            elif line:
                last_line = line
        return tests


def is_float(value: str) -> bool:
    try:
        float(value)
        return True
    except ValueError:
        return False


RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Run tests for Scheme interpreter")
    parser.add_argument("-i", "--interpreter", choices=["python", "c", "typescript", "go"], default="python", help="Interpreter to use")
    args = parser.parse_args()
    
    config = INTERPRETER_CONFIG[args.interpreter]
    if config["compile_command"]:
        subprocess.run(config["compile_command"], shell=True, check=True)
    
    for test_name, test_file in TEST_CONFIG.items():
        print(f"Running test: {test_name}")
        tests = load_tests_from_file(test_file)
        print(f"Run command: {config['run_command']} {test_file}")
        result = subprocess.run(f"{config['run_command']} {test_file}", shell=True, capture_output=True, text=True)
        if result.returncode != 0:
            print(f"{RED}Test {test_name} failed: {result.stderr}{NC}")
            continue
        
        output = result.stdout.strip()
        for i, test in enumerate(tests):
            expected = test.expected
            actual = output.splitlines()[i]
            
            if is_float(expected) and is_float(actual):
                expected = float(expected)
                actual = float(actual)
                
            if expected != actual:
                print(f"{RED}Test {test_name} failed on expression: {test.expression}{NC}")
                print(f"  Expected: {expected}, Actual: {actual}\n")
                break
        else:
            print(f"{GREEN}Test {test_name} passed{NC}")