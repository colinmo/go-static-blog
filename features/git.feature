# https://github.com/cucumber/godog
Feature: Git interaction

    In order to maintain pages in the blog
    As an author
    I need my latest changes in git extracted and sent for conversion

    Scenario Outline: Updating local
        Given the blog and remote are in scenario <scenario>
        When an update is requested with verbose
        Then I should receive <text>

        Examples:
            | scenario | text                                                                             |
            | 1        | No changes found                                                                 |
            | 2        | Create new page "Scenario 1"                                                     |
            | 3        | Update page "Scenario 2"                                                         |
            | 4        | Delete page "Scenario 3"                                                         |
            | 5        | Create new page "Scenario 1"\nUpdate page "Scenario 2"\nDelete page "Scenario 3" |
            | 6        | New image "scenario4.jpg"                                                        |