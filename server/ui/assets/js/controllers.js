var anisyncControllers = angular.module('anisyncControllers', []);

anisyncControllers.controller('AnisyncCtrl', ['$scope', 'Anisync',
  function($scope, Anisync) {
    $scope.check = function(req) {
      $scope.checkResp = Anisync.Check.query(req);
    }
    $scope.sync = function(req) {
      $scope.syncResp = Anisync.Sync.query(req);
    }
  }
]);
