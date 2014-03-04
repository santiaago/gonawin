'use strict';

var tournamentControllers = angular.module('tournamentControllers', []);

tournamentControllers.controller('TournamentListCtrl', ['$scope', 'Tournament', '$location', function($scope, Tournament, $location) {
  console.log('Tournament list controller');
  $scope.tournaments = Tournament.query();

  $scope.tournaments.$promise.then(function(result){
    if(!$scope.tournaments || ($scope.tournaments && !$scope.tournaments.length))
      $scope.noTournamentsMessage = 'You have no tournaments';
  });

  $scope.searchTournament = function(){
    console.log('TournamentListCtrl: searchTournament');
    console.log('keywords: ', $scope.keywords)
    $location.search('q', $scope.keywords).path('/tournaments/search');
  };

  // experimental: sar
  // start world cup create action 
  $scope.createWorldCup = function(){
    console.log('Creating world cup');
    Tournament.saveWorldCup($scope.tournament,
		    function(tournament) {
		      console.log('World Cup Tournament: ', tournament);
		      $location.path('/tournaments/show/' + tournament.Id);
		    },
		    function(err) {
		      console.log('save failed: ', err.data);
		      $scope.messageDanger = err.data;
		    });
  };
  // end world cup create action
}]);

tournamentControllers.controller('TournamentSearchCtrl', ['$scope', '$routeParams', 'Tournament', '$location', function($scope, $routeParams, Tournament, $location) {
  console.log('Tournament search controller');
  console.log('routeParams: ', $routeParams);
  // get tournaments data result from search query
  $scope.tournamentsData = Tournament.search( {q:$routeParams.q});
  
  $scope.tournamentsData.$promise.then(function(result){
    $scope.tournaments = result.Tournaments;
    $scope.messageInfo = result.MessageInfo;
  });
  
  $scope.query = $routeParams.q;
  // use the isSearching mode to differientiate:
  // no tournaments in app AND no tournaments found using query search
  $scope.isSearching = true;
  $scope.searchTournament = function(){
    console.log('TournamentSearchCtrl: searchTournament');
    console.log('keywords: ', $scope.keywords)
    $location.search('q', $scope.keywords).path('/tournaments/search');
  };
}]);

tournamentControllers.controller('TournamentNewCtrl', ['$scope', 'Tournament', '$location', function($scope, Tournament, $location) {
  console.log('Tournament New controller');
  
  $scope.addTournament = function() {
    Tournament.save($scope.tournament,
		    function(tournament) {
		      $location.path('/tournaments/show/' + tournament.Id);
		    },
		    function(err) {
		      console.log('save failed: ', err.data);
		      $scope.messageDanger = err.data;
		    });
  };
}]);

tournamentControllers.controller('TournamentShowCtrl', ['$scope', '$routeParams', 'Tournament', '$location', '$q', function($scope, $routeParams, Tournament, $location, $q) {
  console.log('Tournament Show controller');
  
  $scope.tournamentData =  Tournament.get({ id:$routeParams.id });
  console.log('tournamentData', $scope.tournamentData);

  // get candidates data from tournament id
  $scope.candidatesData = Tournament.candidates({id:$routeParams.id});

  // do we really need theses lines?
  $scope.candidatesData.$promise.then(function(result){
    $scope.candidates = result.Candidates;
  });

  // #experimental: sar
  // list of tournament groups
  $scope.groupsData = Tournament.groups({id:$routeParams.id});
  // admin function: reset tournament
  $scope.resetTournament = function(){
    Tournament.reset({id:$routeParams.id},
		     function(result){
		       console.log('reset succeed.');
		       $scope.messageInfo = result.MessageInfo;
		       $scope.groupsData.Groups = result.Groups;
		     },
		     function(err){
		       console.log('reset failed: ', err.data);
		       $scope.messageDanger = err.data;
		     });
  };


  $scope.deleteTournament = function() {
    Tournament.delete({ id:$routeParams.id },
		      function(){
			$location.path('/');
		      },
		      function(err) {
			console.log('delete failed: ', err.data);
			$scope.messageDanger = err.data;
		      });
  };
  
  $scope.joinTournament = function(){
    Tournament.join({ id:$routeParams.id }).$promise.then(function(result){
      Tournament.participants({ id:$routeParams.id }).$promise.then(function(participantsResult){
        $scope.tournamentData.Participants = participantsResult.Participants;
        $scope.joinButtonName = 'Leave';
        $scope.joinButtonMethod = $scope.leaveTournament;
      });
    });
  };
  
  $scope.leaveTournament = function(){
    Tournament.leave({ id:$routeParams.id }).$promise.then(function(result){
      Tournament.participants({ id:$routeParams.id }).$promise.then(function(participantsResult){
        $scope.tournamentData.Participants = participantsResult.Participants;
        $scope.joinButtonName = 'Join';
        $scope.joinButtonMethod = $scope.joinTournament;
      });
    });
  };
  
  $scope.joinTournamentAsTeam = function(teamId){
    Tournament.joinAsTeam({id:$routeParams.id, teamId:teamId}).$promise.then(function(result){
      Tournament.get({ id:$routeParams.id }).$promise.then(function(tournamentResult){
        $scope.tournamentData.Teams = tournamentResult.Teams;
        $scope.joinAsTeamButtonName[teamId] = 'Leave';
        $scope.joinAsTeamButtonMethod[teamId] = $scope.leaveTournamentAsTeam;
      });
    });
  };
  
  $scope.leaveTournamentAsTeam = function(teamId){
    Tournament.leaveAsTeam({id:$routeParams.id, teamId:teamId}).$promise.then(function(result){
      Tournament.get({ id:$routeParams.id }).$promise.then(function(tournamentResult){
        $scope.tournamentData.Teams = tournamentResult.Teams;
        $scope.joinAsTeamButtonName[teamId] = 'Join';
        $scope.joinAsTeamButtonMethod[teamId] = $scope.joinTournamentAsTeam;
      });
    });
  };
  
  $scope.isTournamentAdmin = $scope.tournamentData.$promise.then(function(result){
    console.log('tournament is admin ready!');
    if(result.Tournament.AdminId == $scope.currentUser.Id){
      return true;
    }else{
      return false;
    }
  });
  
  // checks if user has joined a tournament
  $scope.joined = $scope.tournamentData.$promise.then(function(result){
    console.log('tournament joined ready!');
    return result.Joined;
  });
  
  $scope.tournamentData.$promise.then(function(tournamentResult){
    var deferred = $q.defer();
    if (tournamentResult.Joined) {
      deferred.resolve('Leave');
    }
    else {
      deferred.resolve('Join');
    }
    return deferred.promise;
  }).then(function(result){
    $scope.joinButtonName = result;
  });
  
  $scope.tournamentData.$promise.then(function(tournamentResult){
    var deferred = $q.defer();
    if (tournamentResult.Joined) {
      deferred.resolve($scope.leaveTournament);
    }
    else {
      deferred.resolve($scope.joinTournament);
    }
    return deferred.promise;
  }).then(function(result){
    $scope.joinButtonMethod = result;
  });
  
  $scope.candidatesData.$promise.then(function(candidatesResult){
    var candidatesLength = 0;
    if(candidatesResult.Candidates){
      candidatesLength = candidatesResult.Candidates.length;
    }
    $scope.joinAsTeamButtonName = new Array(candidatesLength);
    $scope.joinAsTeamButtonMethod = new Array(candidatesLength);
    
    $scope.tournamentData.$promise.then(function(tournamentResult){
      for (var i=0 ; i<candidatesLength; i++)
      {
        if(IsTeamJoined(candidatesResult.Candidates[i].Team.Id, tournamentResult.Teams))
        {
          $scope.joinAsTeamButtonName[candidatesResult.Candidates[i].Team.Id] = 'Leave';
          $scope.joinAsTeamButtonMethod[candidatesResult.Candidates[i].Team.Id] = $scope.leaveTournamentAsTeam;
        } else {
          $scope.joinAsTeamButtonName[candidatesResult.Candidates[i].Team.Id] = 'Join';
          $scope.joinAsTeamButtonMethod[candidatesResult.Candidates[i].Team.Id] = $scope.joinTournamentAsTeam;
        }
      }
    });
  });
  
  var IsTeamJoined = function(teamId, teams) {
    if(!teams) {
      return false;
    }
    for (var i=0 ; i<teams.length; i++){
      if(teams[i].Id == teamId){
        return true;
      }
    }
  };
}]);

tournamentControllers.controller('TournamentEditCtrl', ['$scope', '$routeParams', 'Tournament', '$location',function($scope, $routeParams, Tournament, $location) {
  $scope.tournamentData = Tournament.get({ id:$routeParams.id });
  
  $scope.updateTournament = function() {
    var tournamentData = Tournament.get({ id:$routeParams.id });
    Tournament.update({ id:$routeParams.id }, $scope.tournamentData.Tournament,
		      function(){
			$location.path('/tournaments/show/' + $routeParams.id);
		      },
		      function(err) {
			console.log('update failed: ', err.data);
			$scope.messageDanger = err.data;
		      });
  }
}]);

tournamentControllers.controller('TournamentCalendarCtrl', ['$scope', '$routeParams', 'Tournament', '$location',function($scope, $routeParams, Tournament, $location) {
  console.log('Tournament calendar controller');
  console.log('route params', $routeParams)
  $scope.tournamentData = Tournament.get({ id:$routeParams.id });

  $scope.matchesData = Tournament.calendar({id:$routeParams.id, groupby:$routeParams.groupby});

  $scope.activatePredict = function(matchIdNumber, index, parentIndex){
    console.log('Tournament calendar controller: activate predict:', matchIdNumber);
    $scope.matchesData.Days[parentIndex].Matches[index].wantToPredict = true;
  };

  $scope.predict = function(matchIdNumber, index, parentIndex, result1, result2){
    console.log('Tournament calendar controller: predict:', matchIdNumber);

    $scope.matchesData.Days[parentIndex].Matches[index].wantToPredict = false;
    $scope.matchesData.Days[parentIndex].Matches[index].HasPredict = true;
    
    Tournament.predict({id:$routeParams.id, matchId:matchIdNumber, result1:result1, result2:result2},
		       function(result){
			 console.log('success in setting prediction!');
			 $scope.matchesData.Days[parentIndex].Matches[index].Predict = result.Predict.Result1 + ' - ' + result.Predict.Result2;
			 $scope.messageInfo = result.MessageInfo;
			 console.log('match result: ', result.Predict.Result1 + ' - ' + result.Predict.Result2);
		       },
		       function(err) {
			 console.log('failure setting prediction! ', err.data);
			 $scope.messageDanger = err.data;
		       });
    console.log('match result: ', result1, ' ', result2);

  };  

}]);


// Controller for Admin: update results.
// ToDo: Should only be available if you are admin
tournamentControllers.controller('TournamentSetResultsCtrl', ['$scope', '$routeParams', 'Tournament', '$location',function($scope, $routeParams, Tournament, $location) {
  console.log('Tournament set results controller');
  console.log('route params', $routeParams)
  $scope.tournamentData = Tournament.get({ id:$routeParams.id });

  $scope.matchesData = Tournament.calendar({id:$routeParams.id, groupby:"phase"});

  // update result of a match.
  $scope.updateResult = function(match, matchindex, dayindex, phaseindex){
      console.log('TournamentSetResultsCtrl: updateResult');
      console.log('match: ', match);
      console.log('match: ', match.IdNumber);
      console.log('match result: ', match.Result1, ' ', match.Result2);
      console.log('indexes: match, day, phase ', matchindex, dayindex, phaseindex);
      // build result string to send to API
      var result = match.Result1 + ' ' + match.Result2;
      $scope.updatedMatch = Tournament.updateMatchResult({ id:$routeParams.id, matchId:match.IdNumber, result:result});
      // update current match view
      $scope.updatedMatch.$promise.then(function(result){
	  console.log('result: ', result);
	  console.log('matchdata: ', $scope.matchesData);
	  console.log('matchdatamatches: ', $scope.matchesData.Phases[phaseindex].Days[dayindex].Matches[matchindex]);
	  $scope.matchesData.Phases[phaseindex].Days[dayindex].Matches[matchindex] = result;
    });
  };

  // simulate a phase of a tournament.
  $scope.simulatePhase = function(phaseName, phaseindex){
    console.log('TournamentSetResultsCtrl: simulatePhase:', phaseName);
    Tournament.simulatePhase({id:$routeParams.id, phaseName:phaseName},
			     function(result){
			       console.log('success in simulation!');
			       $scope.matchesData.Phases[phaseindex].Days = result.Phase.Days;
			     },
			     function(err) {
			       console.log('failure in  simulation! ', err.data);
			       $scope.messageDanger = err.data;
			     });
  };
    
}]);

tournamentControllers.controller('TournamentFirstStageCtrl',  ['$scope', '$routeParams', 'Tournament', '$location',function($scope, $routeParams, Tournament, $location) {
  console.log('Tournament first stage controller');
  $scope.tournamentData = Tournament.get({ id:$routeParams.id });

  // #experimental: sar
  // list of tournament groups
  $scope.groupsData = Tournament.groups({id:$routeParams.id});
  // predicate is udate for ranking tables
  $scope.predicate = '';

}]);

tournamentControllers.controller('TournamentSecondStageCtrl',  ['$scope', '$routeParams', 'Tournament', '$location',function($scope, $routeParams, Tournament, $location) {
  console.log('Tournament second stage controller');
  $scope.tournamentData = Tournament.get({ id:$routeParams.id });
  $scope.matchesData = Tournament.matches({id:$routeParams.id, filter:"second"});
}]);

// Predict controller
tournamentControllers.controller('TournamentPredictCtrl', ['$scope', '$routeParams', 'Tournament', '$location',function($scope, $routeParams, Tournament, $location) {
  console.log('Tournament predict controller');
  console.log('route params', $routeParams)
  $scope.tournamentData = Tournament.get({ id:$routeParams.id });

  $scope.matchesData = Tournament.calendar({id:$routeParams.id, groupby:$routeParams.groupby});

  $scope.activatePredict = function(matchIdNumber, index, parentIndex){
    console.log('TournamentPredictCtrl: activate predict:', matchIdNumber);
    $scope.matchesData.Days[parentIndex].Matches[index].wantToPredict = true;
  };

  $scope.predict = function(matchIdNumber, index, parentIndex, result1, result2){
    console.log('TournamentPredictCtrl: predict:', matchIdNumber);

    $scope.matchesData.Days[parentIndex].Matches[index].wantToPredict = false;
    $scope.matchesData.Days[parentIndex].Matches[index].HasPredict = true;
    
    Tournament.predict({id:$routeParams.id, matchId:matchIdNumber, result1:result1, result2:result2},
		       function(result){
			 console.log('success in setting prediction!');
			 $scope.matchesData.Days[parentIndex].Matches[index].Predict = result.Predict.Result1 + ' - ' + result.Predict.Result2;
			 $scope.messageInfo = result.MessageInfo;
			 console.log('match result: ', result.Predict.Result1 + ' - ' + result.Predict.Result2);
		       },
		       function(err) {
			 console.log('failure setting prediction! ', err.data);
			 $scope.messageDanger = err.data;
		       });
  };  
}]);

// Ranking controller
tournamentControllers.controller('TournamentRankingCtrl', ['$scope', '$routeParams', 'Tournament', '$location',function($scope, $routeParams, Tournament, $location) {
  console.log('Tournament ranking controller');
  console.log('route params', $routeParams)
  $scope.tournamentData = Tournament.get({ id:$routeParams.id });

  $scope.rankingData = Tournament.ranking({id:$routeParams.id, rankby:$routeParams.rankby});
  // predicate is udate for ranking tables
  $scope.predicate = '';

}]);

