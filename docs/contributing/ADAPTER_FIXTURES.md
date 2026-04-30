# Adapter Fixture Checklist

Every adapter should include fixtures for:

- simple function
- method or class member
- import/module detection
- generated test path
- framework list and default framework
- unsupported file extension
- known false-positive syntax

For mature adapters, add fixtures for async behavior, generics, exceptions, decorators/annotations, and nested classes where the language supports them.
