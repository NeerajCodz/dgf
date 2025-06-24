# Contributing to DGF (Direct Git Fetch)

Thank you for your interest in contributing to DGF! This project is a command-line tool for downloading files and folders from Git repositories, and we welcome contributions to improve its functionality, performance, and usability.

## How to Contribute

### Reporting Issues

- Check the issue tracker to ensure the bug or feature request hasn’t already been reported.
- Open a new issue with a clear title and description, including steps to reproduce (for bugs) or a detailed proposal (for features).

### Suggesting Features

- Propose new features or enhancements via the issue tracker.
- Provide a clear use case and potential implementation details to help us evaluate your suggestion.

### Submitting Code Changes

#### Fork the Repository

- Fork the DGF repository to your GitHub account.
- Clone your fork:

    ```sh
    git clone https://github.com/NeerajCodz/dgf.git
    cd dgf
    ```

#### Create a Branch

- Create a new branch for your changes:

    ```sh
    git checkout -b feature/your-feature-name
    ```
    or
    ```sh
    git checkout -b bugfix/issue-number
    ```

#### Make Changes

- Follow the coding guidelines below.
- Ensure your changes align with the project’s goals (e.g., improving file download functionality or adding support for new platforms).

#### Test Your Changes

- Test your changes locally to ensure they work as expected.
- Verify that existing functionality is not broken.
- If applicable, add tests to cover your changes.

#### Commit Your Changes

- Write clear, concise commit messages:

    ```sh
    git commit -m "Add feature: describe your change"
    ```

- Follow the Conventional Commits format, e.g., `feat: add support for new file format`, `fix: resolve path validation bug`.

#### Push to Your Fork

- Push your branch to your forked repository:

    ```sh
    git push origin feature/your-feature-name
    ```

#### Open a Pull Request

- Go to the DGF repository and open a pull request from your branch.
- Provide a detailed description of your changes, including the problem solved or feature added.
- Reference any related issues (e.g., `Closes #123`).
- Ensure your pull request passes any automated checks (if set up).

## Coding Guidelines

- **Code Style:** Follow the existing code style in the repository. If no specific style guide exists, use consistent formatting (e.g., 2 spaces for indentation in shell scripts).
- **Documentation:** Update relevant documentation (e.g., `README.md` or inline comments) for any new features or changes.
- **Error Handling:** Ensure robust error handling, especially for network operations or file parsing.
- **Performance:** Optimize code for speed and resource usage, as DGF is designed for efficient file downloads.
- **File Formats:** If adding support for new file formats, update the format list in the configuration and document it in the `README.md`.

## Development Setup

- Install DGF using the instructions in the `README.md`.
- Ensure you have the necessary tools (e.g., bash, wget, or curl) to test the installer and the tool.
- [Add any additional setup steps, e.g., specific dependencies or environment setup, if known.]

## Code of Conduct

Please adhere to our Code of Conduct to maintain a respectful and inclusive community.

## Contact

For questions or clarification, reach out via the issue tracker or contact Neeraj SathishKumar via GitHub.

Thank you for contributing to DGF!