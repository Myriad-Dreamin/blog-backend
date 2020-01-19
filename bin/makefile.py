#!/usr/bin/env python3
import os
import re
import subprocess
import sys

made = set()


def consume_makefile(method, *arg, **kwarg):
    if callable(method):
        i = str(method)
        if i not in made:
            made.add(i)
            method(*arg, **kwarg)


# return function in object scope
def sq(obj, attr):
    # raise if not exists
    def __get_object_attr(*_): return getattr(obj, attr)

    return __get_object_attr


# object requirements
def sqs(obj, *attrs): return map(lambda attr: sqs(obj, attr), attrs)


# return function in object scope
def oq(attr):
    # raise if not exists
    def __get_object_attr(obj, *_): return getattr(obj, attr)

    return __get_object_attr


# object requirements
def oqs(*attrs): return map(oq, attrs)


def require(*targets):
    def c(f):
        def wrapper(*args, **kwargs):
            for target in targets:
                target = (getattr(target, '__name__', None) == '__get_object_attr' and target(*args,
                                                                                              **kwargs)) or target
                consume_makefile(target, *args, **kwargs)
            return f(*args, **kwargs)

        return wrapper

    return c


def require_cls(*target): return require(*oqs(*target))


def pipe(cmd, cwd=None):
    p, line, line2 = subprocess.Popen(cmd, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                                      cwd=cwd), ' ', ' '
    while len(line) != 0 or len(line2) != 0:
        line = p.stdout.readline()
        line2 = p.stderr.readline()
        if len(line) != 0:
            print(line.decode('utf-8').strip())
        if len(line2) != 0:
            print(line2.decode('utf-8').strip())
    code = p.wait()
    if code != 0:
        print('exit with %d' % code)


def entry(self, args):
    getattr(self, args[0])(*args[1:]) if len(args) > 0 else self.all()


class Makefile:
    current_path = os.path.dirname(os.path.realpath(__file__))
    config_file_name = 'config.toml'
    context = dict()

    silent = False

    @classmethod
    def hello(cls, *_):
        print('minimum builder')

    @classmethod
    def pipe(cls, cmd, *_):
        if not cls.silent: print(cmd)
        pipe(cmd)

    @classmethod
    def generate(cls, path='./', match=None, *_):
        match = cls._gen_match(match)
        for file in os.listdir(path):
            file = os.path.join(path, file)
            if os.path.isdir(file):
                cls.generate(file, match)
            if os.path.isfile(file) and match.match(file):
                cls.pipe('go generate %s' % file)

    @staticmethod
    def _gen_match(match):
        if match is None:
            match = re.compile(r'^.*\.go$')
        if isinstance(match, str):
            match = re.compile(match)
        return match

    @classmethod
    @require_cls('up')
    def all(cls, *_):
        pass


if __name__ == '__main__':
    entry(Makefile, sys.argv[1:])
