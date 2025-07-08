# FGA Migration

## How to update the FGA model

1. Create a new FGA module file named `<module-name>.fga` containing your authorization model definitions
2. Update the `fga.mod` file to include the new module in the build configuration
3. Execute `./generate.sh` to generate the required output files:
   - `fga.json` - JSON representation of the FGA model used by the backend application
   - `fga.fga` - Combined FGA model file for developer review and validation
   - `fga.md5` - MD5 hash of the FGA model used for change detection during migrations, if the md5 hash is different from the existing one stored in the database, the migration will be applied

## Note
- Currently, the authorization models for other backends are centralized and stored here. In the future, we may consider decentralizing these models and distributing them to their respective backends.
