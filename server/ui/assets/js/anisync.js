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
