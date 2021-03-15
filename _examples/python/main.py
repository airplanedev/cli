#!/usr/bin/env python
import click

@click.command()
@click.option('--name', default="World", help='Who to say hello to.')
def main(name):
    """An example Hello World program."""
    click.secho('Hello %s!' % name, fg="green")

if __name__ == '__main__':
    main()
