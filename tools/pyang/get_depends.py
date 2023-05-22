#! /usr/bin/env python3
################################################################################
#                                                                              #
#  Copyright 2023 Broadcom. The term Broadcom refers to Broadcom Inc. and/or   #
#  its subsidiaries.                                                           #
#                                                                              #
#  Licensed under the Apache License, Version 2.0 (the "License");             #
#  you may not use this file except in compliance with the License.            #
#  You may obtain a copy of the License at                                     #
#                                                                              #
#     http://www.apache.org/licenses/LICENSE-2.0                               #
#                                                                              #
#  Unless required by applicable law or agreed to in writing, software         #
#  distributed under the License is distributed on an "AS IS" BASIS,           #
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.    #
#  See the License for the specific language governing permissions and         #
#  limitations under the License.                                              #
#                                                                              #
################################################################################

import argparse
import pathlib

import pyang
if pyang.__version__ > '2.4':
    from pyang.repository import FileRepository
    from pyang.context import Context
else:
    from pyang import FileRepository
    from pyang import Context


def process(args):
    yangpath = str(args.path.absolute())
    repo = FileRepository(yangpath, use_env=True)
    ctx = Context(repo)
    modules = {}
    for f in args.files:
        f = args.path.joinpath(f)
        if not f.exists():
            continue
        m = ctx.add_module(f.stem, f.read_text())
        modules[f.stem] = m
    ctx.validate()

    imports = set()
    excludes = [name for name in modules]
    if args.exclude and args.exclude.is_dir():
        excludes += [f.stem for f in args.exclude.glob("**/sonic-*.yang")]

    def collect_imports(mod):
        if mod is None:
            return
        unused = [m.arg for m in mod.i_unused_prefixes.values()]
        for i in mod.search("import"):
            if args.exclude_unused and i.arg in unused:
                continue
            if i.arg in excludes:
                continue
            imports.add(i.arg)
            collect_imports(ctx.get_module(i.arg))

    for mod in modules.values():
        collect_imports(mod)

    for name in imports:
        y = args.path.joinpath(name + ".yang")
        if y.exists():
            print(f"{y.name}")


if __name__ == "__main__":
    ap = argparse.ArgumentParser(description="Find dependent yangs")
    ap.add_argument("-p", "--path", help="Yang search directory", type=pathlib.Path)
    ap.add_argument("-x", "--exclude", help="Exclude yang modules which are already present here", type=pathlib.Path)
    ap.add_argument("--exclude-unused", help="Exclude unused imports", action="store_true")
    ap.add_argument("files", help="Yang file names", nargs="+", metavar="FILENAME")
    args = ap.parse_args()
    process(args)
