# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

#### Added - 30-07-2021
- Fix array object field parsing

#### Added - 12-06-2021
- Refactor schema parse into segregation of basic and custom type

#### Added - 10-06-2021
- Segregate Parser into different modules ie apis,schema, goMod, operations etc
- Create interfaces for dependency injection in parser for testing

#### Added - 7-06-2021
- Add yaml spec generation feature

#### Added - 05-06-2021
- Add parsing of type spec from list of alias
- Add type spec schema parsing from multiple alias import

#### Changed - 02-06-2021
- Done some fixes in parsing JSON schema
- An added feature to skip an attribute from struct
- Added Header struct parsing done via comment
- Code module separation done
- Enum header variable parsing from comments done
- Enum Param parsing
- Parse Parameters
- Fix interface type parsing from response comments
- Added logrous for more structured logging.
- Add characterization test to cover some basic scenarios.
- Add a flag schema-without-pkg to create the schema without pkg names
- Add "override-example" tag for scenarios where you want to override a field type example
- Update readMe
- Create parse schema interface and inject in parser
- Add unit test for parse header params

### Merged Changes
 - https://github.com/mikunalpha/goas
