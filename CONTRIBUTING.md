# Contributing to HashPost

Thank you for your interest in contributing to HashPost! We welcome contributions that align with our vision of privacy-focused, community-driven social spaces.

## A message from the maintainer

I'll drop the royal "we" briefly; this project is entirely proprietary. I have not decided on a license, if I will license, or when I will license. I cannot commit to any form of you, as a contributor, retaining or regaining your rights if you submit code to HashPost. Most of what follows is me trying to be extra-clear so that no one feels burned if I never open up the licensing for HashPost.

If you want to work on HashPost today because it's a cool project trying to do some interesting things then great -- I'm happy to have you. If losing your rights or proprietary software are out of the question then I hope you'll be a user when HashPost launches an alpha one day.

Also, frankly, this project may go nowhere. I started this project to see if I could combine code generation and LLM agents to produce robust, well tested code with minimal staff (just me!) and I've gotten as far as what you can see. This is my third such attempt and although I've smoothed out the process greatly I still have many mountains, mole hills, and challenges to overcome with this process.

Maybe lastly, if you think AI or LLMs in general only generate slop and that detests you then that's cool too. I hope you'll be open to providing useful input on what can be better that I didn't catch.

## Important Notice: Proprietary Software

HashPost is proprietary software at the moment and there is no immediate plan to change that. Users must accept that all rights are reserved and that by contributing they transfer the rights of the code to HashPost. We expect that users will create forks in order to make changes but outside of development, contributions, or research they may not use HashPost's code.

## How to Contribute

### Before You Start

1. **Read the Vision**: Familiarize yourself with HashPost's vision and principles in the [README](README.md)
2. **Check Documentation**: Review the [technical documentation](docs/) to understand the codebase
3. **Open an Issue**: Discuss your proposed contribution before starting work

### Development Setup

1. **Fork the Repository**: Create a fork of the HashPost repository on GitHub
2. **Clone Your Fork**: `git clone https://github.com/your-username/hashpost.git`
3. **Set Up Development Environment**: Follow the [development documentation](docs/development.md)
4. **Create a Branch**: `git checkout -b feature/your-feature-name`
5. **Make Your Changes**: Implement your feature or fix
6. **Test Your Changes**: Ensure all tests pass and the application runs correctly

**Note**: While you can fork the repository to contribute, you may not use, run, or distribute the code due to the proprietary license. The fork is solely for the purpose of creating pull requests.

### Code Standards

#### General Guidelines
- Follow Go best practices for backend code
- Use TypeScript for frontend development
- Write comprehensive tests for new functionality
- Update documentation for any API changes
- Ensure all tests pass before submitting

#### Commit Messages
- Use clear, descriptive commit messages
- Start with a verb (Add, Fix, Update, etc.)
- Keep the first line under 50 characters
- Add details in the body if needed

#### Testing
- Write unit tests for new functionality
- Use the isolated test framework for integration tests
- Ensure test coverage for security-critical features
- Run the full test suite before submitting

### Submitting Contributions

1. **Test Your Changes**: Ensure all tests pass and the application runs correctly
2. **Update Documentation**: Add or update relevant documentation
3. **Create a Pull Request**: Submit your changes with a clear description
4. **Respond to Feedback**: Be open to suggestions and improvements

### Pull Request Guidelines

#### What to Include
- Clear description of the changes
- Link to related issues
- Screenshots for UI changes
- Test results and coverage information
- Any breaking changes or migration notes

#### Review Process
- All contributions require review
- Security-related changes require additional review
- We may request changes or improvements
- Final approval is at our discretion