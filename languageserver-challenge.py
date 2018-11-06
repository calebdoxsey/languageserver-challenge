import sublime
import sublime_plugin


import json

import os
import sys
sys.path.append(os.path.join(os.path.dirname(__file__), "vendor"))

import requests


class FindFirstReferenceCommand(sublime_plugin.TextCommand):
    def run(self, edit):
        # TODO(cbd): what to do if there is no file path?
        file_path = self.view.window().extract_variables().get("file", "")
        row, col = self.view.rowcol(self.view.sel()[0].begin())

        req = {
            "FilePath": file_path,
            "Row": row + 1,  # go is 1-based for the parser
            "Column": col + 1,
        }
        res = requests.post('http://127.0.0.1:5000/findreferences',
                            data=json.dumps(req)).json()

        print("response: {}".format(res))

        if not res.get("OK"):
            # do nothing
            return

        for ref in res.get("References", []):
            dst = "{}:{}:{}".format(ref.get("FilePath"),
                                    ref.get("Row"),
                                    ref.get("Column"))
            self.view.window().open_file(dst, sublime.ENCODED_POSITION)
            return


class GoToDefinitionCommand(sublime_plugin.TextCommand):
    def run(self, edit):
        # TODO(cbd): what to do if there is no file path?
        file_path = self.view.window().extract_variables().get("file", "")
        row, col = self.view.rowcol(self.view.sel()[0].begin())

        req = {
            "FilePath": file_path,
            "Row": row + 1,  # go is 1-based for the parser
            "Column": col + 1,
        }
        res = requests.post('http://127.0.0.1:5000/gotodefinition',
                            data=json.dumps(req)).json()

        print("response: {}".format(res))

        if not res.get("OK"):
            # do nothing
            return

        dst = "{}:{}:{}".format(res.get("FilePath"),
                                res.get("Row"),
                                res.get("Column"))
        self.view.window().open_file(dst, sublime.ENCODED_POSITION)


class HoverView(sublime_plugin.ViewEventListener):
    def __init__(self, view):
        self.view = view

    @classmethod
    def is_applicable(cls, settings):
        syntax = settings.get('syntax')
        return syntax == 'Packages/Go/Go.sublime-syntax'

    def on_hover(self, point, hover_zone):
        file_path = self.view.window().extract_variables().get("file", "")
        row, col = self.view.rowcol(point)

        req = {
            "FilePath": file_path,
            "Row": row + 1,  # go is 1-based for the parser
            "Column": col + 1,
        }
        res = requests.post('http://127.0.0.1:5000/hover',
                            data=json.dumps(req)).json()

        print("response: {}".format(res))

        if not res.get("OK"):
            self.view.hide_popup()
            return

        self.view.show_popup("<p>" + res.get("Text") + "</p>",
                             flags=sublime.HIDE_ON_MOUSE_MOVE_AWAY,
                             max_width=300,
                             max_height=300)
