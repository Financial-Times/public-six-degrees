class public_six-degrees {

  $configParameters = hiera('configParameters','')

  class { "go_service_profile" :
    service_module => $module_name,
    service_name => 'public-six-degrees',
    configParameters => $configParameters
  }

}
