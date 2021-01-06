# Challenge.gov scraper

[Challenge.gov](https://www.challenge.gov/) is a government website that hosts
prize competitions and challenges across the U.S. federal government.

All of the currently active challenges listed are on the homepage, with more
details in permalinks for each.

AFAICT, there is no RSS feed or way to be notified where there new challenges
posted. So this project is a way to get the challenges into a machine-readable
format by scraping the homepage periodically.

## Usage

This repo is set up to work as an automated, periodic process in the manner of
[Git scraping as described by Simon
Willison](https://simonwillison.net/2020/Oct/9/git-scraping/). See
`.github/workflows/scraper.yml`.
