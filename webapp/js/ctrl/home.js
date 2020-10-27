// Home controller
plik.controller('HomeCtrl', ['$scope', '$api', '$config', '$dialog', '$location',
    function ($scope, $api, $config, $dialog, $location) {

        $scope.display = 'uploads';
        $scope.displayUploads = function (token) {
            $scope.uploads = [];
            $scope.token = token;
            $scope.display = 'uploads';
            $scope.refreshUser();
        };

        $scope.displayTokens = function () {
            $scope.display = 'tokens';
            $scope.refreshUser();
        };

        $scope.displayInvites = function () {
            $scope.display = 'invites';
            $scope.refreshUser();
        };

        $scope.displayInviteLink = function(invite) {
            var url = window.location.origin + "/#/register?invite=" + invite.id;
            if (invite.email) {
                url += "&email=" + invite.email;
            }
            $dialog.alert(url);
        };

        // Get server config
        $config.config
            .then(function (config) {
                // Check if authentication is enabled server side
                if (!config.authentication) {
                    $location.path('/');
                }
                $scope.config = config;
            })
            .then(null, function (error) {
                $dialog.alert(error);
            });

        // Handle user promise
        var loadUser = function (promise) {
            promise.then(function (user) {
                if (!user.verified) {
                    $location.path('/confirm');
                    return;
                }
                $scope.user = user;
                $scope.getUploads();
                $scope.getTokens();
                $scope.getInvites();
                $scope.getUserStats();
            })
            .then(null, function (error) {
                if (error.status === 401 || error.status === 403) {
                    $location.path('/login');
                } else {
                    $dialog.alert(error);
                }
            });
        };

        // Refresh user
        $scope.refreshUser = function () {
            loadUser($config.refreshUser());
        };

        // page size
        $scope.limit = 50;

        // Get user upload list
        $scope.getUploads = function (more) {
            if (!more) {
                $scope.uploads = [];
                $scope.uploads_cursor = undefined;
            }

            // Get user uploads
            $api.getUserUploads($scope.token, $scope.limit, $scope.uploads_cursor)
                .then(function (result) {
                    $scope.uploads = $scope.uploads.concat(result.results);
                    $scope.uploads_cursor = result.after;
                })
                .then(null, function (error) {
                    $dialog.alert(error);
                });
        };

        // Get user tokens
        $scope.getTokens = function (more) {
            if (!more) {
                $scope.tokens = [];
                $scope.tokens_cursor = undefined;
            }

            // Get user uploads
            $api.getUserTokens($scope.limit, $scope.tokens_cursor)
                .then(function (result) {
                    $scope.tokens = $scope.tokens.concat(result.results);
                    $scope.tokens_cursor = result.after;
                })
                .then(null, function (error) {
                    $dialog.alert(error);
                });
        };

        // Get user invites
        $scope.getInvites = function (more) {
            if (!more) {
                $scope.invites = [];
                $scope.invites_cursor = undefined;
            }

            // Get user uploads
            $api.getUserInvites($scope.limit, $scope.invites_cursor)
                .then(function (result) {
                    $scope.invites = $scope.invites.concat(result.results);
                    $scope.invites_cursor = result.after;
                })
                .then(null, function (error) {
                    $dialog.alert(error);
                });
        };

        // Get user statistics
        $scope.getUserStats = function () {
            $api.getUserStats()
                .then(function (stats) {
                    $scope.user.stats = stats;
                })
                .then(null, function (error) {
                    $dialog.alert(error);
                });
        };

        // Remove an upload
        $scope.deleteUpload = function (upload) {
            $dialog.alert({
                title: "Really ?",
                message: "This will remove " + upload.files.length + " file(s) from the server",
                confirm: true
            }).result.then(
                function () {
                    $api.removeUpload(upload)
                        .then(function () {
                            $scope.uploads = _.reject($scope.uploads, function (u) {
                                return u.id === upload.id;
                            });
                        })
                        .then(null, function (error) {
                            $dialog.alert(error);
                        });
                }, function () {
                    // Avoid "Possibly unhandled rejection"
                });
        };

        // Delete all user uploads
        $scope.deleteUploads = function () {
            $dialog.alert({
                title: "Really ?",
                message: "This will remove all uploads from the server",
                confirm: true
            }).result.then(
                function () {
                    $api.deleteUploads($scope.token)
                        .then(function (result) {
                            $scope.uploads = [];
                            $scope.getUploads();
                            $dialog.alert(result);
                        })
                        .then(null, function (error) {
                            $dialog.alert(error);
                        });
                }, function () {
                    // Avoid "Possibly unhandled rejection"
                });
        };

        // Generate a new token
        $scope.createToken = function (comment) {
            $api.createToken(comment)
                .then(function () {
                    $scope.refreshUser();
                })
                .then(null, function (error) {
                    $dialog.alert(error);
                });
        };

        // Revoke a token
        $scope.revokeToken = function (token) {
            $dialog.alert({
                title: "Really ?",
                message: "Revoking a token will not delete associated uploads.",
                confirm: true
            }).result.then(
                function () {
                    $api.revokeToken(token.token)
                        .then(function () {
                            $scope.refreshUser();
                        })
                        .then(null, function (error) {
                            $dialog.alert(error);
                        });
                }, function () {
                    // Avoid "Possibly unhandled rejection"
                });
        };

        // Generate a new invite
        $scope.createInvite = function (email) {
            $api.createInvite(email)
                .then(function () {
                    $scope.refreshUser();
                })
                .then(null, function (error) {
                    $dialog.alert(error);
                });
        };

        // Revoke a invite
        $scope.revokeInvite = function (invite) {
            $dialog.alert({
                title: "Really ?",
                message: "",
                confirm: true
            }).result.then(
                function () {
                    $api.revokeInvite(invite.id)
                        .then(function () {
                            $scope.refreshUser();
                        })
                        .then(null, function (error) {
                            $dialog.alert(error);
                        });
                }, function () {
                    // Avoid "Possibly unhandled rejection"
                });
        };


        // Log out
        $scope.logout = function () {
            $api.logout()
                .then(function () {
                    $config.refreshUser();
                    $location.path('/');
                })
                .then(null, function (error) {
                    $dialog.alert(error);
                });
        };

        // Sign out
        $scope.deleteAccount = function () {
            $dialog.alert({
                title: "Really ?",
                message: "Deleting your account will not delete your uploads.",
                confirm: true
            }).result.then(
                function () {
                    $api.deleteAccount()
                        .then(function () {
                            $config.refreshUser();
                            $location.path('/');
                        })
                        .then(null, function (error) {
                            $dialog.alert(error);
                        });
                }, function () {
                    // Avoid "Possibly unhandled rejection"
                }
            );
        };

        // Get upload url
        $scope.getUploadUrl = function (upload) {
            return $api.base + '/#/?id=' + upload.id;
        };

        // Get file url
        $scope.getFileUrl = function (upload, file) {
            return $api.base + '/file/' + upload.id + '/' + file.id + '/' + file.fileName;
        };

        // Compute human readable size
        $scope.humanReadableSize = function (size) {
            if (_.isUndefined(size)) return;
            return filesize(size, {base: 2});
        };

        // Redirect to main page
        $scope.mainpage = function () {
            $location.search({});
            $location.hash("");
            $location.path('/');
        };

        loadUser($config.getUser());
    }]);