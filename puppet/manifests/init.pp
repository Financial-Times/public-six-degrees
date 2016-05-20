class public_six-degrees_api {

  $configParameters = hiera('configParameters','')

  class { "go_service_profile" :
    service_module => $module_name,
    service_name => 'public-six-degrees-api',
    configParameters => $configParameters
  }

}
