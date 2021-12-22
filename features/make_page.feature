# https://github.com/cucumber/godog
Feature: Making a page

    In order to maintain pages in the blog
    As an author
    I need to have my git pages converted into html pages

    Scenario Outline: Making pages
        Given the page <scenario> exists
        When I run create page <scenariopage>
        Then I should see a file <file> with contents <scenario>

        Examples:
            | scenario | scenariopage | file         |
            | basic    | today.md     | today.html   |
            | rename   | tomorrow.md  | funday.html  |
            | gallery  | gallery.md   | gallery.html |