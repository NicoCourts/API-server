## NicoCourts.com Backend API Server
![travis-ci status dev](https://travis-ci.com/NicoCourts/API-server.svg?branch=Live)

This is a mini server written in Go that handles requests from the frontend as well as provding a central repository for the blog posts.

# Plans
I would like this to provide a JSON interface to be read by the front end and to provide an API for creating new posts and reading/updating old ones that uses cryptographic signatures to provide authentication for sensitive routes (e.g. updating and deleting)

My original plan was to use `pandoc` or something similar on the server side to handle document conversion, but that has ended up being largely irrelevant since I want to do the conversion from markdown to LaTeX-enriched HTML on the client side anyways to enable previews.
