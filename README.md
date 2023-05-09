# mattermost-timeout-users-plugin
This is a plugin for Mattermost that prevents users from using channel wide mentions in Mattermost servers. This way a
server admin can configure the acceptable threshold of `@channel`, `@all` or `@here` and the timeout of the users.
This is a way of putting users in their place and making them think before posting a message that mentions the whole 
server, as time is money.

## Develop
To develop for this project you will need
*  Make
*  Go
*  Git

To avoid having to manually install your plugin, build and deploy your plugin using one of the following options. In order for the below options to work, you must first enable plugin uploads via your config.json or API and restart Mattermost.

```json
    "PluginSettings" : {
        ...
        "EnableUploads" : true
    }
```


To build the plugin use the following command
```shell
make
```
This will build the plugin as a single binary for all the supported architectures and then create a
bundle of the plugin that can be used to upload it to a Mattermost server

Then Navigate to System Console > Plugin Management, upload the plugin bundle and configure it.

*Enjoy a life without (many) distractions*
