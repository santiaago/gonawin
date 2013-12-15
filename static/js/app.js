'use strict';

var purpleWingApp = angular.module('purpleWingApp', ['ngSanitize', 'directive.g+signin', 'ngCookies'])
    .config(function($routeProvider){
	$routeProvider.when('/',
			    {
				templateUrl: 'templates/main.html', 
				controller: 'MainController'
			    });
	$routeProvider.when('/about',
			    {
				templateUrl: 'templates/about.html'
			    });
	$routeProvider.when('/contact',
			    {
				templateUrl: 'templates/contact.html'
			    });
	$routeProvider.when('/users/:userId',
			    {
				templateUrl: 'templates/user_show.html',
				controller: 'UserShowController'
			    });
	$routeProvider.when('/teams',
			    {
				templateUrl: 'templates/teams.html',
				controller: 'TeamsController'
			    });
	$routeProvider.when('/teams/new',
			    {
				templateUrl: 'templates/teams_new.html',
				controller: 'TeamsNewController'
			    });
	$routeProvider.when('/teams/:teamId',
			    {
				templateUrl: 'templates/teams_show.html',
				controller: 'TeamsShowController'
			    });
	$routeProvider.when('/settings/edit-profile',
			    {
				templateUrl: 'templates/settings_edit-profile.html',
				controller: 'SettingsEditProfileController'
			    });
	$routeProvider.when('/settings/networks',
			    {
				templateUrl: 'templates/settings_networks.html'
			    });
	$routeProvider.when('/settings/email',
			    {
				templateUrl: 'templates/settings_email.html'
			    });

	$routeProvider.when('/invite',
			    {
				templateUrl: 'templates/invite.html',
				controller: 'InviteController'
			    });

	$routeProvider.otherwise( {redirectTo: '/'});
    })
    .factory('myCache', function($cacheFactory){
	return $cacheFactory('myCache', {capacity:3})
    });
