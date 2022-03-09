var assert = require('assert');

(function(){
    'use strict';
    
    equal('should pass', function() {
        assert(1 === 1);
      });
      
    function equal(desc, fn) {
      try {
        fn();
        console.log('\x1b[32m%s\x1b[0m', '\u2714 ' + desc);
        console.log("Add your tests in this ./test directory");
      } catch (error) {
        console.log('\n');
        console.log('\x1b[31m%s\x1b[0m', '\u2718 ' + desc);
        console.error(error);
      }
    }
  })();