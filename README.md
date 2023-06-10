# VNS-scraper

It's a tool to pass silly tests in VNS automatically. 
**Attention!!!** You can only use this tool for tests with open correct answers after completion and unlimited attempts.

## Installation
```
go install github.com/pryamcem/vns-scraper@latest
```

## How to use it
```
vns-scraper pass [link to a test from VNS]
```
Automatically passing the test and storing the correct answers to storage.

```
vns-scraper save [number of test to save]
```
Saving the completed test from storage to a text file.

```
vns-scraper scan [link to a test from VNS]
```
If you have the passed test by yourself and you want to save the correct answers to storage, you can use this command.

## How it works
It uses [go-rod](https://github.com/go-rod/rod) library which uses [Chrome DevTools Protocol](https://chromedevtools.github.io/devtools-protocol/) to render web pages. When the page was rendered it starts a new attempt at a test. It scans all questions and tries to find answers from storage and clicks to It on a test.  If the test was unsuccessful it scans all right answers and stores them in storage. Then starts a new attempt while the test will not complete successfully.
