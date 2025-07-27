/**
 * Version switcher for Globus Go SDK documentation
 */
document.addEventListener('DOMContentLoaded', function() {
  // Get version selector element
  const versionSelector = document.querySelector('.version-selector select');
  if (!versionSelector) return;
  
  // Save the current path within the version
  let currentPath = window.location.pathname;
  
  // Get all available versions from the data attribute
  const availableVersions = JSON.parse(
    document.querySelector('meta[name="available-versions"]').getAttribute('content')
  );
  
  // Extract the version from the current path
  let currentVersion = '';
  for (const version of availableVersions) {
    if (currentPath.includes(version.path)) {
      currentVersion = version.version;
      // Remove version path from current path to get relative path
      currentPath = currentPath.replace(version.path, '');
      break;
    }
  }
  
  // Set up version change handler
  versionSelector.addEventListener('change', function(e) {
    // Get the selected version path
    const selectedVersionPath = e.target.value;
    
    // Construct the new URL with the relative path
    let newPath = selectedVersionPath;
    if (currentPath && currentPath !== '/') {
      // Normalize path to avoid double slashes
      newPath = `${selectedVersionPath.replace(/\/$/, '')}${currentPath}`;
    }
    
    // Navigate to the new version path
    window.location.href = newPath;
  });
  
  // Show version information on API elements if available
  const versionTagElements = document.querySelectorAll('[data-api-version]');
  versionTagElements.forEach(function(element) {
    const apiVersion = element.getAttribute('data-api-version');
    const deprecatedInVersion = element.getAttribute('data-deprecated-in');
    const addedInVersion = element.getAttribute('data-added-in');
    
    if (apiVersion) {
      const versionTag = document.createElement('span');
      versionTag.classList.add('api-version-tag');
      versionTag.textContent = `API ${apiVersion}`;
      
      if (deprecatedInVersion) {
        versionTag.classList.add('deprecated');
        versionTag.title = `Deprecated in v${deprecatedInVersion}`;
      } else if (addedInVersion) {
        versionTag.classList.add('new');
        versionTag.title = `Added in v${addedInVersion}`;
      }
      
      // Insert at the beginning of the element
      element.insertBefore(versionTag, element.firstChild);
    }
  });
});