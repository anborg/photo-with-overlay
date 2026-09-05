//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreLocation -framework Foundation

#import <CoreLocation/CoreLocation.h>
#import <Foundation/Foundation.h>
#import <dispatch/dispatch.h>
#import <stdlib.h>

typedef struct {
	double latitude;
	double longitude;
	double accuracy;
	char* error;
} NativeLocationResult;

@interface PhotoWithOverlayLocationDelegate : NSObject<CLLocationManagerDelegate>
@property(nonatomic, strong) CLLocationManager* manager;
@property(nonatomic, strong) CLLocation* location;
@property(nonatomic, strong) NSError* error;
@property(nonatomic, strong) dispatch_semaphore_t semaphore;
@end

@implementation PhotoWithOverlayLocationDelegate

- (void)locationManagerDidChangeAuthorization:(CLLocationManager *)manager {
	CLAuthorizationStatus status;
	if (@available(macOS 11.0, *)) {
		status = manager.authorizationStatus;
	} else {
		status = [CLLocationManager authorizationStatus];
	}
	if (status == kCLAuthorizationStatusAuthorized || status == kCLAuthorizationStatusAuthorizedAlways) {
		[manager startUpdatingLocation];
		return;
	}
	if (status == kCLAuthorizationStatusDenied || status == kCLAuthorizationStatusRestricted) {
		self.error = [NSError errorWithDomain:@"PhotoWithOverlayLocation"
			code: 1
			userInfo:@{NSLocalizedDescriptionKey: @"Location access denied by macOS"}];
		dispatch_semaphore_signal(self.semaphore);
	}
}

- (void)locationManager:(CLLocationManager *)manager didUpdateLocations:(NSArray<CLLocation *> *)locations {
	CLLocation* last = [locations lastObject];
	if (last != nil) {
		self.location = last;
		[manager stopUpdatingLocation];
		dispatch_semaphore_signal(self.semaphore);
	}
}

- (void)locationManager:(CLLocationManager *)manager didFailWithError:(NSError *)error {
	self.error = error;
	[manager stopUpdatingLocation];
	dispatch_semaphore_signal(self.semaphore);
}

@end

static NativeLocationResult PhotoWithOverlayRequestCurrentLocation(double timeoutSeconds) {
	NativeLocationResult result;
	result.latitude = 0;
	result.longitude = 0;
	result.accuracy = -1;
	result.error = NULL;

	if (![CLLocationManager locationServicesEnabled]) {
		result.error = strdup("Location Services is disabled on this Mac");
		return result;
	}

	__block PhotoWithOverlayLocationDelegate* delegate = nil;
	void (^startRequest)(void) = ^{
		delegate = [PhotoWithOverlayLocationDelegate new];
		delegate.semaphore = dispatch_semaphore_create(0);
		delegate.manager = [CLLocationManager new];
		delegate.manager.delegate = delegate;
		delegate.manager.desiredAccuracy = kCLLocationAccuracyBest;

		CLAuthorizationStatus status;
		if (@available(macOS 11.0, *)) {
			status = delegate.manager.authorizationStatus;
		} else {
			status = [CLLocationManager authorizationStatus];
		}

		if (status == kCLAuthorizationStatusNotDetermined) {
			[delegate.manager requestWhenInUseAuthorization];
		} else if (status == kCLAuthorizationStatusAuthorized || status == kCLAuthorizationStatusAuthorizedAlways) {
			[delegate.manager startUpdatingLocation];
		} else {
			delegate.error = [NSError errorWithDomain:@"PhotoWithOverlayLocation"
				code: 1
				userInfo:@{NSLocalizedDescriptionKey: @"Location access denied by macOS"}];
			dispatch_semaphore_signal(delegate.semaphore);
		}
	};

	if ([NSThread isMainThread]) {
		startRequest();
	} else {
		dispatch_sync(dispatch_get_main_queue(), startRequest);
	}

	long waitResult = dispatch_semaphore_wait(delegate.semaphore, dispatch_time(DISPATCH_TIME_NOW, (int64_t)(timeoutSeconds * NSEC_PER_SEC)));
	if (waitResult != 0) {
		if ([NSThread isMainThread]) {
			[delegate.manager stopUpdatingLocation];
		} else {
			dispatch_sync(dispatch_get_main_queue(), ^{
				[delegate.manager stopUpdatingLocation];
			});
		}
		result.error = strdup("Timed out waiting for macOS location");
		return result;
	}

	if (delegate.error != nil) {
		result.error = strdup([[delegate.error localizedDescription] UTF8String]);
		return result;
	}
	if (delegate.location == nil) {
		result.error = strdup("macOS did not return a location");
		return result;
	}

	result.latitude = delegate.location.coordinate.latitude;
	result.longitude = delegate.location.coordinate.longitude;
	result.accuracy = delegate.location.horizontalAccuracy;
	return result;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func getCurrentLocation() (CurrentLocation, error) {
	result := C.PhotoWithOverlayRequestCurrentLocation(20)
	if result.error != nil {
		defer C.free(unsafe.Pointer(result.error))
		return CurrentLocation{}, fmt.Errorf("%s", C.GoString(result.error))
	}

	location := CurrentLocation{
		Latitude:  float64(result.latitude),
		Longitude: float64(result.longitude),
	}
	if result.accuracy >= 0 {
		accuracy := float64(result.accuracy)
		location.Accuracy = &accuracy
	}
	return location, nil
}
