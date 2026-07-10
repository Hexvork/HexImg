cmake_minimum_required(VERSION 3.21)

project(HexImg VERSION 0.2.0 LANGUAGES CXX)

include(CTest)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_AUTOMOC ON)
set(CMAKE_AUTORCC ON)
set(CMAKE_AUTOUIC ON)

find_package(Qt6 6.5 REQUIRED COMPONENTS Core Gui Qml Quick QuickControls2 Svg Widgets)

qt_add_executable(HexImg WIN32
    src/main.cpp
    src/heximgbackend.cpp
    src/heximgbackend.h
    src/queuemodel.cpp
    src/queuemodel.h
)

qt_add_resources(HexImg "qml"
    PREFIX "/"
    FILES
        qml/Main.qml
        qml/FormatBadge.qml
        qml/ShadButton.qml
        qml/ShadComboBox.qml
        qml/ShadTextField.qml
        assets/HexImg.png
)

target_link_libraries(HexImg PRIVATE
    Qt6::Core
    Qt6::Gui
    Qt6::Qml
    Qt6::Quick
    Qt6::QuickControls2
    Qt6::Svg
    Qt6::Widgets
)

if(WIN32)
    target_sources(HexImg PRIVATE packaging/windows/HexImg.rc)
    target_link_libraries(HexImg PRIVATE ole32 windowscodecs)
endif()

if(WIN32)
    set_target_properties(HexImg PROPERTIES WIN32_EXECUTABLE TRUE)
endif()

install(TARGETS HexImg
    BUNDLE DESTINATION .
    RUNTIME DESTINATION .
)

if(BUILD_TESTING)
    find_package(Qt6 6.5 REQUIRED COMPONENTS Test)

    qt_add_executable(HexImgTests
        tests/heximgbackend_test.cpp
        src/heximgbackend.cpp
        src/heximgbackend.h
        src/queuemodel.cpp
        src/queuemodel.h
    )
    target_include_directories(HexImgTests PRIVATE src)
    target_link_libraries(HexImgTests PRIVATE
        Qt6::Core
        Qt6::Gui
        Qt6::Svg
        Qt6::Test
        Qt6::Widgets
    )

    if(WIN32)
        target_link_libraries(HexImgTests PRIVATE ole32 windowscodecs)
    endif()

    add_test(NAME HexImgTests COMMAND HexImgTests -platform offscreen)
endif()
