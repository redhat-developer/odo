---
layout: default
permalink: /tdd-workflow/
redirect_from: 
  - /docs/tdd-workflow.md/
---

## Test Driven Development (TDD) for odo

What’s TDD? In simple terms, WRITE TESTS FIRST.

#### TDD Workflow

Under TDD workflow, we write tests before writing feature code. Below is the workflow that is generally followed:

1. Add a test
2. Run all tests and see if the new test fails
3. Write the code
4. Run tests
5. Rewrite code incorporating any changes, and add new tests

#### TDD process for odo

1. Write tests for every feature/bug fix before implementing the code.
	* Developer will split the assigned task into small achievable subtasks
	* Each subtask will require unit tests to be written before feature implementation
	* Open work-in-progress [WIP] PR with the tests written

2. The tests should cover both positive and negative scenarios:
	* It would be good to have all permutations and combinations covered *as much as possible* helping us cover the boundary cases
	* There is always scope to add  more tests could be added once the feature has been fully implemented

3. Write feature code and run tests:
	* Feature code should be pushed to the same WIP test PR
	* Feature code should pushed only after pushing the unit tests
	* After the feature code is pushed, the test should pass

4. Review process
	* Ensure that PR with code change is accompanied by unit tests
	* Travis checks to be our point of determination

#### FAQs
1. ##### Won't TDD slow down our development process?

   Initially it might feel that way, but over time, we will realise the advantage this would bring to our development process by reducing the bugs, improving test coverage and the development time spent fixing regression bugs.

2. ##### What if we don't know what our code function will finally look like

   TDD is just a recommendation. Nothing is stopping you writing your code before your test (locally, of course!).

3. ##### Writing tests based on a new feature is easier.

    Writing tests beforehand (under TDD) would help us identify the design limitations and corner cases in advance. Writing tests post development often limits us to cover positive cases. TDD will help us move away from that mindset.

4. ##### Can we follow the same process for big features too?

    Tasks involving big features should be split into numerous simple and achievable subtasks, and each of those subtasks should have test coverage.

5. ##### Should tests always be written first?

    That’s the point. In order to follow TDD process, we need to look at feature development from a testing perspective. It will help us cover wider scenarios beyond positive cases. Tests will also serve as good validator during the coding process.

#### Learning resources

If you haven't tried TDD before, here are a few resources to get started!

1. https://github.com/quii/learn-go-with-tests
2. https://cacoo.com/blog/test-driven-development-in-go/
