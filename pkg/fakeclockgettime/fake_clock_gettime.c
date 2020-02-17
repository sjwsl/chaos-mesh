#include <time.h>
#include <inttypes.h>

int64_t TV_SEC_DELTA = 0;
int64_t TV_NSEC_DELTA = 0;

long syscall(long number, ...);

int clock_gettime(clockid_t clk_id, struct timespec *tp) {
    int ret;
    asm volatile
        (
            "syscall"
            : "=a" (ret)
            : "0"(228), "D"(clk_id), "S"(tp)
            : "rcx", "r11", "memory"
        );

    int64_t sec_delta = TV_SEC_DELTA;
    int64_t nsec_delta = TV_NSEC_DELTA;

    int64_t billion = 1000000000;

    if(clk_id == CLOCK_REALTIME) {
        while (nsec_delta + tp->tv_nsec > billion) {
            sec_delta += 1;
            nsec_delta -= billion;
        }

        while (nsec_delta + tp->tv_nsec < 0) {
            sec_delta -= 1;
            nsec_delta += billion;
        }

        tp->tv_sec += sec_delta;
        tp->tv_nsec += nsec_delta;
    }

    return ret;
}