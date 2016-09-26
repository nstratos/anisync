var anisyncApp = angular.module('anisyncApp', ['ngMdIcons', 'anisyncControllers', 'anisyncServices', 'ngProgress', 'remoteValidation']);

var anisyncServices = angular.module('anisyncServices', ['ngResource']);

anisyncServices.factory('Anisync', ['$resource',
  function($resource) {
    var checkURL = 'api/check';
    var syncURL = 'api/sync';
    if (window.location.search == "?testbed") {
      checkURL = 'api/mock/check';
      syncURL = 'api/mock/sync';
    }
    return {
      Check: $resource(checkURL, {}, {
        query: {
          method: 'GET',
          params: {}
        },
      }),
      Sync: $resource(syncURL, {}, {
        query: {
          method: 'POST',
          params: {}
        },
      })
    }
  }
]);

anisyncServices.factory('Sync', ['$resource', ]);

var anisyncControllers = angular.module('anisyncControllers', []);

anisyncControllers.controller('AnisyncCtrl', ['$scope', 'Anisync', 'ngProgressFactory', '$window', '$timeout',

  function($scope, Anisync, ngProgressFactory, $window, $timeout) {
    // malVerify options for the password input
    $scope.malVerifyURL = '/api/mal-verify';
    $scope.malVerifyDelay = 900;
    if (window.location.search == "?testbed") {
      $scope.malVerifyURL = '/api/mock/mal-verify';
      $scope.malVerifyDelay = 400;
    }
    // clickNext
    $scope.clickNext = function() {
      $scope.switchOn = true;
      $scope.turnSwitchOn();
      $timeout(function() {
        $window.document.getElementById("malPasswordInput").focus();
      });
    };
    // turnSwitchOn
    $scope.turnSwitchOn = function() {
      $scope.mainForm.malPassword.$setPristine();
      $scope.mainForm.malPassword.$setUntouched();
      angular.element(document.querySelector('#malPasswordInput')).val("");
      angular.element(document.querySelector('#syncButton')).attr("disabled", "disabled");
    };
    // turnSwitchOff
    $scope.turnSwitchOff = function() {
      $scope.mainForm.$setValidity("ngRemoteValidate", undefined);
      $scope.mainForm.malPassword.$setValidity("ngRemoteValidate", undefined);
      $scope.mainForm.$pending = false;
    };
    // updateSwitch
    $scope.updateSwitch = function() {
      if ($scope.switchOn) {
        $scope.turnSwitchOn();
      } else {
        $scope.turnSwitchOff();
      }
    };
    // Modifying ngRemoteValidate for malPassword field so that it
    // sends the username along with the password value.
    $scope.malPasswordSetArgs = function(val, el, attrs, ngModel) {
      return {
        malPassword: val,
        malUsername: $scope.req.malUsername
      };
    };
    // Initializing ngProgress bar.
    $scope.progressbar = ngProgressFactory.createInstance();
    $scope.progressbar.setColor('#ec8661');
    $scope.check = function(req) {
      if ($scope.statusBar) $scope.statusBar.visible = false;
      $scope.statusBar = {};
      $scope.checkResp = {};
      $scope.progressbar.start();
      $scope.loading = true;
      Anisync.Check.query(req).$promise.then(function(data) {
        $scope.checkResp = data;
        $scope.statusBar = makeStatusBar(data);
      }, function(error) {
        $scope.statusBar = makeStatusBarError(error);
      }).finally(function() {
        $scope.progressbar.complete();
        $scope.loading = false;
      });
    }
    $scope.sync = function(req) {
      $scope.statusBar = {};
      $scope.checkResp = {};
      $scope.progressbar.start();
      $scope.loading = true;
      Anisync.Sync.query(req).$promise.then(function(data) {
        $scope.checkResp = data;
        //$scope.syncResp = data;
        $scope.statusBar = makeStatusBarSync(data);
      }, function(error) {
        $scope.statusBar = makeStatusBarError(error);
      }).finally(function() {
        $scope.progressbar.complete();
        $scope.loading = false;
      });
    }
  }
]);

function makeStatusBarError(error) {
  var statusBar = {
    type: "simple",
    message: "",
    cause: "",
    visible: true,
    next: false,
    askMessage: "",
    theme: "danger",
    buttonTheme: "danger"
  };
  if (!error) {
    statusBar.message = "Aw, Snap! Something went horribly wrong.";
    return statusBar;
  }
  statusBar.message = error.data.Message;
  statusBar.cause = error.data.Cause;
  return statusBar;
}

function makeStatusBarSync(data) {
  var statusBar = {
    type: "simple",
    message: "",
    errorMessage: "",
    visible: true,
    next: false,
    details: true,
    detailsVisible: false,
    askMessage: "Sync",
    theme: "success"
  };
  statusBar.message += "MyAnimeList.net account \"" + data.MalUsername + "\" just had:\n";

  if (data.Sync.Adds && data.Sync.Updates) {
    statusBar.message += data.Sync.Updates.length + " updated and " + data.Sync.Adds.length + " newly added anime.";
  }
  if (data.Sync.Adds && !data.Sync.Updates) {
    statusBar.message += data.Sync.Adds.length + " newly added anime.";
  }
  if (!data.Sync.Adds && data.Sync.Updates) {
    statusBar.message += data.Sync.Updates.length + " updated anime.";
  }
  if (data.Sync.AddFails || data.Sync.UpdateFails) {
    statusBar.theme = "danger";
    statusBar.buttonTheme = "danger";
    statusBar.type = "details";
  }
  if (data.Sync.AddFails && data.Sync.UpdateFails) {
    statusBar.message += " However " + data.Sync.UpdateFails.length +
      " failed to update and " + data.Sync.AddFails.length + " failed to be added.";
  }
  if (!data.Sync.AddFails && data.Sync.UpdateFails) {
    statusBar.message += " However " + data.Sync.UpdateFails.length + " failed to update.";
  }
  if (data.Sync.AddFails && !data.Sync.UpdateFails) {
    statusBar.message += " However " + data.Sync.AddFails.length + " failed to be added.";
  }

  if (data.Sync.UpdateFails) {
    statusBar.errorMessage += "Update failures:\n";
    for (i = 0; i < data.Sync.UpdateFails.length; i++) {
      statusBar.errorMessage += "Name: " + data.Sync.UpdateFails[i].Anime.Title + ", Reason: " + data.Sync.UpdateFails[i].Reason + "\n";
    }
  }
  if (data.Sync.AddFails) {
    statusBar.errorMessage += "Add failures:\n";
    for (i = 0; i < data.Sync.AddFails.length; i++) {
      statusBar.errorMessage += "Name: " + data.Sync.AddFails[i].Anime.Title + ", Reason: " + data.Sync.AddFails[i].Reason + "\n";
    }
  }
  return statusBar;
}

function makeStatusBar(data) {
  var statusBar = {
    type: "next",
    message: "",
    visible: true,
    next: true,
    askMessage: "Sync",
    theme: "info",
    buttonTheme: "success"
  };
  // create status message
  statusBar.message = "After syncing, there will be:\n"
  if (data.Missing && data.NeedUpdate) {
    statusBar.message += data.NeedUpdate.length + " updated and " + data.Missing.length + " newly added ";
  }
  if (data.Missing && !data.NeedUpdate) {
    statusBar.message += data.Missing.length + " newly added ";
  }
  if (!data.Missing && data.NeedUpdate) {
    statusBar.message += data.NeedUpdate.length + " updated ";
  }
  statusBar.message += "anime on MyAnimeList.net account \"" + data.MalUsername + "\".";
  if (!data.Missing && !data.NeedUpdate) {
    statusBar.message = "Everything is in sync!";
    statusBar.theme = "success";
    statusBar.next = false;
    statusBar.type = "simple";
  }
  return statusBar;
}
