var anisyncApp = angular.module('anisyncApp', ['ngMdIcons', 'anisyncControllers', 'anisyncServices', 'ngProgress', 'remoteValidation']);

var anisyncServices = angular.module('anisyncServices', ['ngResource']);

anisyncServices.factory('Anisync', ['$resource',
  function($resource) {
    return {
      Check: $resource('api/test/check', {}, {
        query: {
          method: 'GET',
          params: {}
        }
      }),
      Sync: $resource('api/sync', {}, {
        query: {
          method: 'POST',
          params: {}
        }
      })
    }
  }
]);

anisyncServices.factory('Sync', ['$resource', ]);

var anisyncControllers = angular.module('anisyncControllers', []);

anisyncControllers.controller('AnisyncCtrl', ['$scope', 'Anisync', 'ngProgressFactory',
  function($scope, Anisync, ngProgressFactory) {
    $scope.switchToggle = function() {
      $scope.switchOn = !$scope.switchOn;

      if ($scope.switchOn == false) {
        $scope.mainForm.$setValidity("ngRemoteValidate", undefined);
        $scope.mainForm.malPassword.$setValidity("ngRemoteValidate", undefined);
        $scope.mainForm.$pending = false;
      }
      if ($scope.switchOn == true) {
        $scope.mainForm.malPassword.$setPristine();
        $scope.mainForm.malPassword.$setUntouched();
        angular.element(document.querySelector('#malPasswordInput')).val("");
        angular.element(document.querySelector('#syncButton')).attr("disabled", "disabled");
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
      $scope.checkResp = {};
      $scope.progressbar.start();
      $scope.loading = true;
      Anisync.Check.query(req).$promise.then(function(data) {
        $scope.checkResp = data;
        $scope.statusBar = makeStatusBar(data);
        console.log(data);
        console.log($scope);
      }).finally(function() {
        $scope.progressbar.complete();
        $scope.loading = false;
      });
    }
    $scope.sync = function(req) {
      $scope.syncResp = {};
      $scope.progressbar.start();
      $scope.loading = true;
      Anisync.Sync.query(req).$promise.then(function(data) {
        $scope.syncResp = data;
      }).finally(function() {
        $scope.progressbar.complete();
        $scope.loading = false;
      });
    }
  }
]);

function makeStatusBar(data) {
  var statusBar = {
    message: "",
    visible: true,
    next: true,
    askMessage: "Sync",
    type: "info"
  };
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
    statusBar.type = "success";
    statusBar.next = false;
  }
  return statusBar;
}
