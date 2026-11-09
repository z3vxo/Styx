#pragma once
#include "../agent/core.h"



namespace utils {
	namespace mem {
		inline int memcmp( const void* s1, const void* s2, SIZE_T n )
		{
			const volatile unsigned char* p1 = ( const volatile unsigned char * ) s1;
			const volatile unsigned char* p2 = ( const volatile unsigned char * ) s2;

			for ( SIZE_T i = 0; i < n; i++ ) {
				if ( p1 [ i ] != p2 [ i ] ) {
					return p1 [ i ] - p2 [ i ];
				}
			}
			return 0;
		}

		inline void* memcpy(void* dst, const void* src, SIZE_T n)
		{
			volatile unsigned char* d = ( volatile unsigned char * ) dst;
			const volatile unsigned char* s = ( const volatile unsigned char * ) src;
			while ( n-- ) *d++  = *s++;
			return dst;
		}

		inline void* memset( void* dst, int val, SIZE_T n )
		{
			volatile unsigned char* d = ( volatile unsigned char * ) dst;
			while ( n-- ) *d++ = ( unsigned char )val;
			return dst;
		}
	}



	namespace str {
		inline SIZE_T strlen( const char* str )
		{
			const volatile char* s = ( const volatile char * )str;
			SIZE_T n = 0;
			while ( *s++ ) n++;
			return n;
		}

		inline int strcmp( const char* s1, const char* s2 ) {
			while ( *s1 && ( *s1 == *s2 )) {
				s1++;
				s2++;
			}

			return ( unsigned char )*s1 - ( unsigned char )*s2;
		}

		inline int strncmp( const char* s1, const char* s2, SIZE_T n )
		{
			const volatile unsigned char* p1 = ( const volatile unsigned char * )s1;
			const volatile unsigned char* p2 = ( const volatile unsigned char * )s2;
			for ( SIZE_T i = 0; i < n; i++ ) {
				if ( p1 [ i ] != p2 [ i ] )
					return p1 [ i ] - p2 [ i ];
				if ( p1 [ i ] == '\0')
					return 0;
			}
			return 0;
		}

		inline char* strchr( const char* str, int c ) {
			volatile char* s = (volatile char * )str;
			char ch = ( char )c;
			while ( *s != ch ) {
				if ( *s == '\0' )
					return NULL;
				s++;
			}
			return ( char * )s;
		}
	}


	namespace win {
		inline PPEB GetPeb() {
#if defined(_WIN64) || defined(__x86_64__)
			return ( PPEB )__readgsqword( 0x60 );
#elif defined(_M_IX86)|| defined(__i386__)
			return ( PPEB )__readfsdword( 0x30 );
#else
			return NULL;
#endif
		}

		inline PTEB GetTeb() {
#if defined(_WIN64) || defined(__x86_64__)
			return ( PTEB )__readgsqword( 0x30 );
#elif defined(_M_IX86)|| defined(__i386__)
			return ( PTEB )__readfsdword( 0x18 );
#else
			return NULL;
#endif
		}
	}
}

