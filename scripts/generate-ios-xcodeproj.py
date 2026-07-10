#!/usr/bin/env python3
"""Generate a minimal Plexus.xcodeproj for apps/ios/Plexus."""

from __future__ import annotations

import hashlib
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1] / "apps" / "ios" / "Plexus"
PROJECT_DIR = ROOT / "Plexus.xcodeproj"
SCHEME_DIR = PROJECT_DIR / "xcshareddata" / "xcschemes"

SOURCE_FILES = [
    "App/PlexusApp.swift",
    "App/MainTabView.swift",
    "App/RootView.swift",
    "Features/Auth/LoginView.swift",
    "Features/Auth/RegisterView.swift",
    "Features/Backlog/BacklogView.swift",
    "Features/Board/BoardView.swift",
    "Features/Issues/CreateIssueSheet.swift",
    "Features/Issues/IssueDetailView.swift",
    "Features/Projects/CreateProjectSheet.swift",
    "Features/Workspace/CreateWorkspaceSheet.swift",
    "Shared/Models/Models.swift",
    "Shared/Network/APIClient.swift",
    "Shared/Stores/AuthStore.swift",
    "Shared/Stores/KeychainStore.swift",
]


def uid(seed: str) -> str:
    return hashlib.md5(f"plexus-ios-{seed}".encode()).hexdigest()[:24].upper()


IDS = {
    "project": uid("project"),
    "target": uid("target"),
    "product": uid("product"),
    "sources_phase": uid("sources"),
    "project_config_list": uid("project-config-list"),
    "target_config_list": uid("target-config-list"),
    "project_debug": uid("project-debug"),
    "project_release": uid("project-release"),
    "target_debug": uid("target-debug"),
    "target_release": uid("target-release"),
    "main_group": uid("main-group"),
    "products_group": uid("products-group"),
    "plexus_group": uid("plexus-group"),
    "info_plist": uid("info-plist"),
}

for index, path in enumerate(SOURCE_FILES):
    stem = path.replace("/", "-")
    IDS[f"file-{index}"] = uid(f"file-{stem}")
    IDS[f"build-{index}"] = uid(f"build-{stem}")


def pbx_file_ref(file_id: str, path: str, file_type: str = "sourcecode.swift") -> str:
    return (
        f"\t\t{file_id} /* {Path(path).name} */ = {{isa = PBXFileReference; "
        f"lastKnownFileType = {file_type}; path = {path}; sourceTree = \"<group>\"; }};"
    )


def pbx_build_file(build_id: str, file_id: str, name: str) -> str:
    return (
        f"\t\t{build_id} /* {name} in Sources */ = {{isa = PBXBuildFile; "
        f"fileRef = {file_id} /* {name} */; }};"
    )


build_files = [
    pbx_build_file(IDS[f"build-{i}"], IDS[f"file-{i}"], Path(path).name)
    for i, path in enumerate(SOURCE_FILES)
]

file_refs = [
    f'\t\t{IDS["product"]} /* Plexus.app */ = {{isa = PBXFileReference; explicitFileType = wrapper.application; includeInIndex = 0; path = Plexus.app; sourceTree = BUILT_PRODUCTS_DIR; }};',
    f'\t\t{IDS["info_plist"]} /* Info.plist */ = {{isa = PBXFileReference; lastKnownFileType = text.plist.xml; path = Info.plist; sourceTree = "<group>"; }};',
]
file_refs.extend(pbx_file_ref(IDS[f"file-{i}"], path) for i, path in enumerate(SOURCE_FILES))

source_entries = [
    f'\t\t\t\t{IDS[f"build-{i}"]} /* {Path(path).name} in Sources */,'
    for i, path in enumerate(SOURCE_FILES)
]

groups = f"""
\t\t{IDS["main_group"]} = {{
\t\t\tisa = PBXGroup;
\t\t\tchildren = (
\t\t\t\t{IDS["plexus_group"]} /* Plexus */,
\t\t\t\t{IDS["products_group"]} /* Products */,
\t\t\t);
\t\t\tsourceTree = "<group>";
\t\t}};
\t\t{IDS["products_group"]} /* Products */ = {{
\t\t\tisa = PBXGroup;
\t\t\tchildren = (
\t\t\t\t{IDS["product"]} /* Plexus.app */,
\t\t\t);
\t\t\tname = Products;
\t\t\tsourceTree = "<group>";
\t\t}};
\t\t{IDS["plexus_group"]} /* Plexus */ = {{
\t\t\tisa = PBXGroup;
\t\t\tchildren = (
\t\t\t\t{IDS["info_plist"]} /* Info.plist */,
{chr(10).join(f'\t\t\t\t{IDS[f"file-{i}"]} /* {Path(path).name} */,' for i, path in enumerate(SOURCE_FILES))}
\t\t\t);
\t\t\tpath = .;
\t\t\tsourceTree = "<group>";
\t\t}};"""

common_target_settings = """
\t\t\t\tASSETCATALOG_COMPILER_GENERATE_SWIFT_ASSET_SYMBOL_EXTENSIONS = YES;
\t\t\t\tCLANG_ENABLE_MODULES = YES;
\t\t\t\tCODE_SIGN_STYLE = Automatic;
\t\t\t\tCURRENT_PROJECT_VERSION = 1;
\t\t\t\tDEVELOPMENT_TEAM = "";
\t\t\t\tENABLE_PREVIEWS = YES;
\t\t\t\tGENERATE_INFOPLIST_FILE = NO;
\t\t\t\tINFOPLIST_FILE = Info.plist;
\t\t\t\tIPHONEOS_DEPLOYMENT_TARGET = 17.0;
\t\t\t\tLD_RUNPATH_SEARCH_PATHS = (
\t\t\t\t\t"$(inherited)",
\t\t\t\t\t"@executable_path/Frameworks",
\t\t\t\t);
\t\t\t\tMARKETING_VERSION = 1.0.0;
\t\t\t\tPRODUCT_BUNDLE_IDENTIFIER = app.plexus;
\t\t\t\tPRODUCT_NAME = "$(TARGET_NAME)";
\t\t\t\tSDKROOT = iphoneos;
\t\t\t\tSWIFT_EMIT_LOC_STRINGS = YES;
\t\t\t\tSWIFT_VERSION = 5.0;
\t\t\t\tTARGETED_DEVICE_FAMILY = "1,2";"""

project = f"""// !$*UTF8*$!
{{
\tarchiveVersion = 1;
\tclasses = {{
\t}};
\tobjectVersion = 56;
\tobjects = {{

/* Begin PBXBuildFile section */
{chr(10).join(build_files)}
/* End PBXBuildFile section */

/* Begin PBXFileReference section */
{chr(10).join(file_refs)}
/* End PBXFileReference section */

/* Begin PBXGroup section */
{groups}
/* End PBXGroup section */

/* Begin PBXNativeTarget section */
\t\t{IDS["target"]} /* Plexus */ = {{
\t\t\tisa = PBXNativeTarget;
\t\t\tbuildConfigurationList = {IDS["target_config_list"]} /* Build configuration list for PBXNativeTarget "Plexus" */;
\t\t\tbuildPhases = (
\t\t\t\t{IDS["sources_phase"]} /* Sources */,
\t\t\t);
\t\t\tbuildRules = (
\t\t\t);
\t\t\tdependencies = (
\t\t\t);
\t\t\tname = Plexus;
\t\t\tproductName = Plexus;
\t\t\tproductReference = {IDS["product"]} /* Plexus.app */;
\t\t\tproductType = "com.apple.product-type.application";
\t\t}};
/* End PBXNativeTarget section */

/* Begin PBXProject section */
\t\t{IDS["project"]} /* Project object */ = {{
\t\t\tisa = PBXProject;
\t\t\tattributes = {{
\t\t\t\tBuildIndependentTargetsInParallel = 1;
\t\t\t\tLastSwiftUpdateCheck = 1600;
\t\t\t\tLastUpgradeCheck = 1600;
\t\t\t\tTargetAttributes = {{
\t\t\t\t\t{IDS["target"]} = {{
\t\t\t\t\t\tCreatedOnToolsVersion = 16.0;
\t\t\t\t\t}};
\t\t\t\t}};
\t\t\t}};
\t\t\tbuildConfigurationList = {IDS["project_config_list"]} /* Build configuration list for PBXProject "Plexus" */;
\t\t\tcompatibilityVersion = "Xcode 14.0";
\t\t\tdevelopmentRegion = en;
\t\t\thasScannedForEncodings = 0;
\t\t\tknownRegions = (
\t\t\t\ten,
\t\t\t\tBase,
\t\t\t);
\t\t\tmainGroup = {IDS["main_group"]};
\t\t\tproductRefGroup = {IDS["products_group"]} /* Products */;
\t\t\tprojectDirPath = "";
\t\t\tprojectRoot = "";
\t\t\ttargets = (
\t\t\t\t{IDS["target"]} /* Plexus */,
\t\t\t);
\t\t}};
/* End PBXProject section */

/* Begin PBXSourcesBuildPhase section */
\t\t{IDS["sources_phase"]} /* Sources */ = {{
\t\t\tisa = PBXSourcesBuildPhase;
\t\t\tbuildActionMask = 2147483647;
\t\t\tfiles = (
{chr(10).join(source_entries)}
\t\t\t);
\t\t\trunOnlyForDeploymentPostprocessing = 0;
\t\t}};
/* End PBXSourcesBuildPhase section */

/* Begin XCBuildConfiguration section */
\t\t{IDS["project_debug"]} /* Debug */ = {{
\t\t\tisa = XCBuildConfiguration;
\t\t\tbuildSettings = {{
\t\t\t\tALWAYS_SEARCH_USER_PATHS = NO;
\t\t\t\tCLANG_ANALYZER_NONNULL = YES;
\t\t\t\tCLANG_ENABLE_MODULES = YES;
\t\t\t\tCLANG_ENABLE_OBJC_ARC = YES;
\t\t\t\tCOPY_PHASE_STRIP = NO;
\t\t\t\tDEBUG_INFORMATION_FORMAT = dwarf;
\t\t\t\tENABLE_STRICT_OBJC_MSGSEND = YES;
\t\t\t\tENABLE_TESTABILITY = YES;
\t\t\t\tGCC_DYNAMIC_NO_PIC = NO;
\t\t\t\tGCC_OPTIMIZATION_LEVEL = 0;
\t\t\t\tGCC_PREPROCESSOR_DEFINITIONS = (
\t\t\t\t\t"DEBUG=1",
\t\t\t\t\t"$(inherited)",
\t\t\t\t);
\t\t\t\tIPHONEOS_DEPLOYMENT_TARGET = 17.0;
\t\t\t\tMTL_ENABLE_DEBUG_INFO = INCLUDE_SOURCE;
\t\t\t\tONLY_ACTIVE_ARCH = YES;
\t\t\t\tSDKROOT = iphoneos;
\t\t\t\tSWIFT_ACTIVE_COMPILATION_CONDITIONS = "DEBUG $(inherited)";
\t\t\t\tSWIFT_OPTIMIZATION_LEVEL = "-Onone";
\t\t\t}};
\t\t\tname = Debug;
\t\t}};
\t\t{IDS["project_release"]} /* Release */ = {{
\t\t\tisa = XCBuildConfiguration;
\t\t\tbuildSettings = {{
\t\t\t\tALWAYS_SEARCH_USER_PATHS = NO;
\t\t\t\tCLANG_ANALYZER_NONNULL = YES;
\t\t\t\tCLANG_ENABLE_MODULES = YES;
\t\t\t\tCLANG_ENABLE_OBJC_ARC = YES;
\t\t\t\tCOPY_PHASE_STRIP = NO;
\t\t\t\tDEBUG_INFORMATION_FORMAT = "dwarf-with-dsym";
\t\t\t\tENABLE_NS_ASSERTIONS = NO;
\t\t\t\tENABLE_STRICT_OBJC_MSGSEND = YES;
\t\t\t\tGCC_OPTIMIZATION_LEVEL = s;
\t\t\t\tIPHONEOS_DEPLOYMENT_TARGET = 17.0;
\t\t\t\tMTL_ENABLE_DEBUG_INFO = NO;
\t\t\t\tSDKROOT = iphoneos;
\t\t\t\tSWIFT_COMPILATION_MODE = wholemodule;
\t\t\t\tVALIDATE_PRODUCT = YES;
\t\t\t}};
\t\t\tname = Release;
\t\t}};
\t\t{IDS["target_debug"]} /* Debug */ = {{
\t\t\tisa = XCBuildConfiguration;
\t\t\tbuildSettings = {{{common_target_settings}
\t\t\t}};
\t\t\tname = Debug;
\t\t}};
\t\t{IDS["target_release"]} /* Release */ = {{
\t\t\tisa = XCBuildConfiguration;
\t\t\tbuildSettings = {{{common_target_settings}
\t\t\t\tVALIDATE_PRODUCT = YES;
\t\t\t}};
\t\t\tname = Release;
\t\t}};
/* End XCBuildConfiguration section */

/* Begin XCConfigurationList section */
\t\t{IDS["project_config_list"]} /* Build configuration list for PBXProject "Plexus" */ = {{
\t\t\tisa = XCConfigurationList;
\t\t\tbuildConfigurations = (
\t\t\t\t{IDS["project_debug"]} /* Debug */,
\t\t\t\t{IDS["project_release"]} /* Release */,
\t\t\t);
\t\t\tdefaultConfigurationIsVisible = 0;
\t\t\tdefaultConfigurationName = Release;
\t\t}};
\t\t{IDS["target_config_list"]} /* Build configuration list for PBXNativeTarget "Plexus" */ = {{
\t\t\tisa = XCConfigurationList;
\t\t\tbuildConfigurations = (
\t\t\t\t{IDS["target_debug"]} /* Debug */,
\t\t\t\t{IDS["target_release"]} /* Release */,
\t\t\t);
\t\t\tdefaultConfigurationIsVisible = 0;
\t\t\tdefaultConfigurationName = Release;
\t\t}};
/* End XCConfigurationList section */
\t}};
\trootObject = {IDS["project"]} /* Project object */;
}}
"""

scheme = f"""<?xml version="1.0" encoding="UTF-8"?>
<Scheme
   LastUpgradeVersion = "1600"
   version = "1.7">
   <BuildAction
      parallelizeBuildables = "YES"
      buildImplicitDependencies = "YES">
      <BuildActionEntries>
         <BuildActionEntry
            buildForTesting = "YES"
            buildForRunning = "YES"
            buildForProfiling = "YES"
            buildForArchiving = "YES"
            buildForAnalyzing = "YES">
            <BuildableReference
               BuildableIdentifier = "primary"
               BlueprintIdentifier = "{IDS["target"]}"
               BuildableName = "Plexus.app"
               BlueprintName = "Plexus"
               ReferencedContainer = "container:Plexus.xcodeproj">
            </BuildableReference>
         </BuildActionEntry>
      </BuildActionEntries>
   </BuildAction>
   <TestAction
      buildConfiguration = "Debug"
      selectedDebuggerIdentifier = "Xcode.DebuggerFoundation.Debugger.LLDB"
      selectedLauncherIdentifier = "Xcode.DebuggerFoundation.Launcher.LLDB"
      shouldUseLaunchSchemeArgsEnv = "YES">
   </TestAction>
   <LaunchAction
      buildConfiguration = "Debug"
      selectedDebuggerIdentifier = "Xcode.DebuggerFoundation.Debugger.LLDB"
      selectedLauncherIdentifier = "Xcode.DebuggerFoundation.Launcher.LLDB"
      launchStyle = "0"
      useCustomWorkingDirectory = "NO"
      ignoresPersistentStateOnLaunch = "NO"
      debugDocumentVersioning = "YES"
      debugServiceExtension = "internal"
      allowLocationSimulation = "YES">
      <BuildableProductRunnable
         runnableDebuggingMode = "0">
         <BuildableReference
            BuildableIdentifier = "primary"
            BlueprintIdentifier = "{IDS["target"]}"
            BuildableName = "Plexus.app"
            BlueprintName = "Plexus"
            ReferencedContainer = "container:Plexus.xcodeproj">
         </BuildableReference>
      </BuildableProductRunnable>
   </LaunchAction>
   <ProfileAction
      buildConfiguration = "Release"
      shouldUseLaunchSchemeArgsEnv = "YES"
      savedToolIdentifier = ""
      useCustomWorkingDirectory = "NO"
      debugDocumentVersioning = "YES">
      <BuildableProductRunnable
         runnableDebuggingMode = "0">
         <BuildableReference
            BuildableIdentifier = "primary"
            BlueprintIdentifier = "{IDS["target"]}"
            BuildableName = "Plexus.app"
            BlueprintName = "Plexus"
            ReferencedContainer = "container:Plexus.xcodeproj">
         </BuildableReference>
      </BuildableProductRunnable>
   </ProfileAction>
   <AnalyzeAction
      buildConfiguration = "Debug">
   </AnalyzeAction>
   <ArchiveAction
      buildConfiguration = "Release"
      revealArchiveInOrganizer = "YES">
   </ArchiveAction>
</Scheme>
"""


def main() -> None:
    PROJECT_DIR.mkdir(parents=True, exist_ok=True)
    SCHEME_DIR.mkdir(parents=True, exist_ok=True)
    (PROJECT_DIR / "project.pbxproj").write_text(project)
    (SCHEME_DIR / "Plexus.xcscheme").write_text(scheme)
    print(f"Generated {PROJECT_DIR}")


if __name__ == "__main__":
    main()
