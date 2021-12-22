Feature: RSS Feed

    In order to have syndication
    As an author
    I need the RSS feed to update

    Scenario Outline: RSS Feed updating
        Given the page <scenario> exists
        And the RSS feed <scenario> exists
        When I run create page <scenariopage>
        Then I should see a file feed.xml with contents <scenario>
        # Tag file for scenario
        And I should see a file <scenario> with contents <scenario>

        Examples:
            | scenario | scenariopage |
            | basic    | today.md     |
            | rename   | tomorrow.md  |
            | gallery  | gallery.md   |