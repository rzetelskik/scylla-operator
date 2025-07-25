# -*- coding: utf-8 -*-
from datetime import date

from sphinx_scylladb_theme.utils import multiversion_regex_builder

# -- General configuration

# Add any Sphinx extension module names here, as strings. They can be
# extensions coming with Sphinx (named 'sphinx.ext.*') or your custom
# ones.
extensions = [
    'sphinx.ext.todo',
    'sphinx.ext.mathjax',
    'sphinx.ext.githubpages',
    'sphinx.ext.extlinks',
    'sphinx_scylladb_theme',
    'sphinx_multiversion',
    "sphinx_sitemap",
    "sphinx_design",
    "myst_parser",
]

# The suffix(es) of source filenames.
# You can specify multiple suffix as a list of string:
source_suffix = ['.rst', '.md']
autosectionlabel_prefix_document = True

# The encoding of source files.
#
# source_encoding = 'utf-8-sig'

# The master toctree document.
master_doc = 'index'

# General information about the project.
project = 'Scylla Operator'
copyright = str(date.today().year) + ', ScyllaDB. All rights reserved.'
author = u'Scylla Project Contributors'

# List of patterns, relative to source directory, that match files and
# directories to ignore when looking for source files.
# This patterns also effect to html_static_path and html_extra_path
exclude_patterns = ['_build', 'Thumbs.db', '.DS_Store', 'hack']

# The name of the Pygments (syntax highlighting) style to use.
pygments_style = 'sphinx'

# If true, `todo` and `todoList` produce output, else they produce nothing.
todo_include_todos = True

# -- Options for myst parser

myst_enable_extensions = ["colon_fence", "attrs_inline", "substitution"]
myst_heading_anchors = 6

# DEPRECATION NOTICE
# MyST substitutions work counterintuitively with multiversion docs. Versions specified in the main branch are used for all versions.
# These variables have no effect if set on branches other than master.
# https://github.com/scylladb/scylla-operator/issues/2795
myst_substitutions = {
  "productName": "Scylla Operator",
  "repository": "scylladb/scylla-operator",
  "revision": "master",
  "imageRepository": "docker.io/scylladb/scylla",
  "imageTag": "2025.1.2",
  "enterpriseImageRepository": "docker.io/scylladb/scylla-enterprise",
  "enterpriseImageTag": "2024.1.12",
  "agentVersion": "3.5.0",
}

# -- Options for not found extension

# Template used to render the 404.html generated by this extension.
notfound_template =  '404.html'

# Prefix added to all the URLs generated in the 404 page.
notfound_urls_prefix = ''

# -- Options for redirect extension

# Read a YAML dictionary of redirections and generate an HTML file for each
redirects_file = "./redirections.yaml"

# -- Options for multiversion extension
# Whitelist pattern for tags (set to None to ignore all tags)
TAGS = []
smv_tag_whitelist = multiversion_regex_builder(TAGS)
# Whitelist pattern for branches (set to None to ignore all branches)
BRANCHES = ['master', 'v1.16', 'v1.17', 'v1.18']
# Set which versions are not released yet.
UNSTABLE_VERSIONS = ["master", "v1.18"]
smv_branch_whitelist = multiversion_regex_builder(BRANCHES)
# Defines which version is considered to be the latest stable version.
# Must be listed in smv_tag_whitelist or smv_branch_whitelist.
smv_latest_version = 'v1.17'
smv_rename_latest_version = 'stable'
# Whitelist pattern for remotes (set to None to use local branches only)
smv_remote_whitelist = r"^origin$"
# Pattern for released versions
smv_released_pattern = r'^tags/.*$'
# Format for versioned output directories inside the build directory
smv_outputdir_format = '{ref.name}'

# -- Options for HTML output

# The theme to use for HTML and HTML Help pages.  See the documentation for
# a list of builtin themes.
#
html_theme = 'sphinx_scylladb_theme'

# Theme options are theme-specific and customize the look and feel of a theme
# further.  For a list of options available for each theme, see the
# documentation.
html_theme_options = {
    'conf_py_path': 'docs/source/',
    'default_branch': 'master',
    'hide_edit_this_page_button': 'false',
    "hide_feedback_buttons": 'false',
    'github_repository': 'scylladb/scylla-operator',
    'github_issues_repository': 'scylladb/scylla-operator',
    "versions_unstable": UNSTABLE_VERSIONS,
    "zendesk_tag": 'd8cgbpqrvmemn8ugficex8',
}

# If not None, a 'Last updated on:' timestamp is inserted at every page
# bottom, using the given strftime format.
# The empty string is equivalent to '%b %d, %Y'.
#
html_last_updated_fmt = '%d %B %Y'

# Custom sidebar templates, maps document names to template names.
#
html_sidebars = {'**': ['side-nav.html']}

# Output file base name for HTML help builder.
htmlhelp_basename = 'ScyllaDocumentationdoc'

# URL which points to the root of the HTML documentation.
html_baseurl = 'https://operator.docs.scylladb.com'

# Dictionary of values to pass into the template engine's context for all pages
html_context = {'html_baseurl': html_baseurl}

# Add the _static directory to the static path
html_static_path = ['_static']

# Add custom JavaScript files
html_js_files = ['fix-cards.js']

sitemap_url_scheme = "/stable/{link}"
