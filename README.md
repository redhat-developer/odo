# Introduction

Welcome! So if you're here... you obviously want to write some documentation!

I'll give you a general overview on how documentation is synchronized between the `master` and `gh-pages`.

In theory, you just run `./scripts/sync.sh`  and BAM. Docs are updated. But we'll go into more detail regarding what's happening.

First of all, we have three sources of documentation that we're pulling from:

  - Downstream: OpenShift productized documentation from https://docs.openshift.com
  - Upstream: Odo public documentation from the docs folder https://github.com/openshift/odo/tree/master/docs

Within these sources, we gather the following from our `master` branch:

  - Blog posts: Blog posts come from https://github.com/openshift/odo/tree/master/docs/blog
  - Documentation: Public documentation comes from https://github.com/openshift/odo/tree/master/docs/public
  - File reference: File reference documentation comes from https://github.com/openshift/odo/tree/master/docs/file-reference

# Files and file directories

Here is a general overview of the file structure to be used as reference:

- `/chosen-docs`: List of docs which have been "chosen" to be included on the site. These are all **TEMPLATE** files.
  - `/chosen-docs/openshift-docs` - Documentation from [docs.openshift.com](https://docs.openshift.com)
  - `/chosen-docs/upstream` - Documentation from the `master` branch of [github.com/openshift/odo](https://github.com/openshift/odo)
- `/docs`: Where all the converted documentation goes, this is what is shown on the site.
- `/default.md`: "Default" template when creating a page
- `/index.md`: Where we list all front-page documentation. This creates the "grid" of documentation as shown on the front page of the site.
- `/slate`: Where we use file reference 
- `/_posts`: Where blog posts are located

# How to update / troubleshoot documentation

First all all, know this: **Only documentation listed within `/chosen-docs` will be SYNCED. These have two requirements: 1. Must be markdown, 2. Must be the same name as the corresponding `.adoc` document it will be syncing.**

Each file within `/chosen-docs` are only TEMPLATE files. These are required in order to specify the title of the document, description, etc. 

1. Check the `/chosen-docs` directory for which file needs updating.
2. Find out where the file is located.
  - If the file is located in the `/upstream` folder, update it here: [github.com/openshift/odo/tree/master/docs/public](https://github.com/openshift/odo/tree/master/docs/public)
  - If the file is located in the `/openshift-docs` folder, update it here: [github.com/openshift/openshift-docs/tree/master/cli_reference/developer_cli_odo](https://github.com/openshift/openshift-docs/tree/master/cli_reference/developer_cli_odo)
3. Push a PR to wherever the original document is (either `odo` or `openshift-docs` repo)
4. Once you have updated the files, run the `./script/sync.sh` file
5. Commit and push a PR to the `gh-pages` branch (this branch)

# How to add new documentation

1. Add the document either on the [master branch docs folder](https://github.com/openshift/odo/tree/master/docs/file-reference) or the [openshift-docs repo](https://github.com/openshift/odo/tree/master/docs/file-reference).
2. Copy the `default.md` template to either the:
  - `/chosen-docs/openshift-docs` location if the original adoc is located at https://github.com/openshift/openshift-docs/tree/master/cli_reference/developer_cli_odo
  - or `/chosen-docs/upstream` if the adoc is located at https://github.com/openshift/odo/tree/master/docs/file-reference
3. Make sure that the `default.md` template file you've copied over has the EXACT same filename as the `adoc` minus the `.adoc` extension. Example: `understanding-odo.adoc` would be `understanding-odo.md` in the `/chosen-docs` folder
4. Update the `index.md` file with a "card" of the new document that you have added, or else it will *not* appear on the site.

# Want to write a blog post?

Looking to write a blog post? Here's what you've got to do!

1. Copy the template.md in the [/docs/blog](https://github.com/openshift/odo/tree/master/docs/blog) folder of the master branch.
2. This new file must be in this date + topic format: `YEAR-MONTH-DATE-TOPIC.md` for example: `2021-10-11-odo-300-release.md`
3. Write your blog and push a PR to the `master` branch for people to review.

**NOTES REGARDING ABOUT AUTHOR SECTION:**

Want to add a picture of yourself as an author? Push to the: [/img](https://github.com/openshift/odo/tree/gh-pages/img) directory of `gh-pages` with a 460px x 460px picture of yourself!

# How to update the file reference documentation

We use [slate](https://github.com/slatedocs/slate) to generate beautiful file reference documentation. This is created using the corresponding line in in the `./scripts/sync.sh` script:

```sh
docker run --rm -v $PWD:/usr/src/app/source -w /usr/src/app/source cdrage/slate bundle exec middleman build --clean && cp -r build ../file-reference
```

Which you can then view by opening the: `/file-reference/index.html` file.

**How to update the documentation:**

In order to update the file reference documentation, you must update the [index.md](https://github.com/openshift/odo/blob/master/docs/file-reference/index.md) documentation and then run `./scripts/sync.sh`.

Alternatively, if you would like to preview changes and see what it will look like, you can:

```sh
 $ cd slate/

 # Make changes to the file reference markdown doc
 # this is only for previewing changes. Do NOT use this for updating docs.
 $ vim slate/source/index.html.md

 # Build the documentation
 $ docker run --rm -v $PWD:/usr/src/app/source -w /usr/src/app/source cdrage/slate bundle exec middleman build --clean

 # View it on your favourite browser
 $ google-chrome-stable build/index.html
```

# Bundling the site for releases

In the event of a release, you may have to bundle the site in a `tar.gz` file:

```sh
# Build the site
$ jekyll build

# Copy over the instructions on how to view the site
$ cp site-readme.txt _site/README.txt

# Change into the static html directory and tarball the site
$ cd _site && tar -zcvf site.tar.gz *
```

After following the above instructions, you will have a `site.tar.gz` with a README.txt located in the root directory.

You can now include this `site.tar.gz` in releases.

# Previewing changes to the site / viewing `gh-pages` locally

The site is built using [jekyll](https://jekyllrb.com/) which is the default static site builder on GitHub.

To preview changes / view them locally, run:

```sh
  # Run the site
  $ jekyll serve .

  # View it online
  $ google-chrome-stable localhost:4000
```
